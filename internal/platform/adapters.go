package platform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BaseAdapter struct {
	meta       ProviderMetadata
	httpClient *http.Client
	allowLocal bool
}

func NewBaseAdapter(meta ProviderMetadata, timeout time.Duration) *BaseAdapter {
	return NewBaseAdapterWithOptions(meta, timeout, false)
}

func NewBaseAdapterWithOptions(meta ProviderMetadata, timeout time.Duration, allowLocal bool) *BaseAdapter {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	return &BaseAdapter{
		meta:       meta,
		httpClient: client,
		allowLocal: allowLocal,
	}
}

func (b *BaseAdapter) Metadata() ProviderMetadata {
	return b.meta
}

func (b *BaseAdapter) ValidateCredential(ctx context.Context, secret string) (bool, string, error) {
	if strings.TrimSpace(secret) == "" {
		return false, "empty credential", nil
	}
	return true, "credential format valid", nil
}

func (b *BaseAdapter) HealthCheck(ctx context.Context, secret string) (HealthStatus, error) {
	start := time.Now()
	if b.meta.BaseURL == "" {
		return HealthStatus{
			Healthy:   true,
			Latency:   time.Since(start),
			Message:   "provider online",
			CheckedAt: time.Now().UTC(),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, b.meta.BaseURL, nil)
	if err != nil {
		return HealthStatus{
			Healthy:   false,
			Latency:   time.Since(start),
			Message:   err.Error(),
			CheckedAt: time.Now().UTC(),
		}, err
	}

	b.injectAuth(req, secret)
	resp, err := b.httpClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return HealthStatus{
			Healthy:   false,
			Latency:   latency,
			Message:   err.Error(),
			CheckedAt: time.Now().UTC(),
		}, nil
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode < 500
	return HealthStatus{
		Healthy:   healthy,
		Latency:   latency,
		Message:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		CheckedAt: time.Now().UTC(),
	}, nil
}

func (b *BaseAdapter) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	targetURL := b.meta.BaseURL
	if req.Path != "" {
		if strings.HasPrefix(req.Path, "http://") || strings.HasPrefix(req.Path, "https://") {
			targetURL = req.Path
		} else {
			targetURL = strings.TrimRight(b.meta.BaseURL, "/") + "/" + strings.TrimLeft(req.Path, "/")
		}
	}

	if !b.allowLocal {
		if err := ValidateSSRF(targetURL); err != nil {
			return nil, fmt.Errorf("SSRF violation: %w", err)
		}
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if len(req.QueryParams) > 0 {
		q := u.Query()
		for k, v := range req.QueryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if len(req.Body) > 0 && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	b.injectAuth(httpReq, req.CredentialSecret)

	start := time.Now()
	resp, err := b.httpClient.Do(httpReq)
	latency := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("http execute error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	respHeaders := make(map[string]string)
	for k := range resp.Header {
		respHeaders[k] = resp.Header.Get(k)
	}

	return &ExecutionResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       respBody,
		Latency:    latency,
	}, nil
}

func (b *BaseAdapter) injectAuth(req *http.Request, secret string) {
	if strings.TrimSpace(secret) == "" {
		return
	}
	switch b.meta.AuthType {
	case AuthTypeBearerToken:
		header := b.meta.AuthHeaderName
		if header == "" {
			header = "Authorization"
		}
		prefix := b.meta.AuthHeaderPrefix
		if prefix == "" {
			prefix = "Bearer "
		}
		req.Header.Set(header, prefix+secret)
	case AuthTypeAPIKeyHeader, AuthTypePersonalToken, AuthTypeCustomHeader:
		header := b.meta.AuthHeaderName
		if header == "" {
			header = "x-api-key"
		}
		prefix := b.meta.AuthHeaderPrefix
		req.Header.Set(header, prefix+secret)
	case AuthTypeBasicAuth:
		parts := strings.SplitN(secret, ":", 2)
		if len(parts) == 2 {
			req.SetBasicAuth(parts[0], parts[1])
		} else {
			req.SetBasicAuth(secret, "")
		}
	}
}

func ValidateSSRF(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("forbidden protocol: %s", scheme)
	}

	hostname := strings.ToLower(u.Hostname())
	if hostname == "" {
		return errors.New("empty hostname")
	}

	if hostname == "localhost" || hostname == "0.0.0.0" || hostname == "127.0.0.1" || hostname == "::1" {
		return fmt.Errorf("target host %s is not permitted (loopback)", hostname)
	}

	if strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".internal") || strings.HasSuffix(hostname, ".localhost") {
		return fmt.Errorf("target internal domain %s is not permitted", hostname)
	}

	ip := net.ParseIP(hostname)
	if ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("target IP %s is within private or metadata range", ip.String())
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}
	if ipv4[0] == 10 {
		return true
	}
	if ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31 {
		return true
	}
	if ipv4[0] == 192 && ipv4[1] == 168 {
		return true
	}
	if ipv4[0] == 169 && ipv4[1] == 254 {
		return true
	}
	return false
}
