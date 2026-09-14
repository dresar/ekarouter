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
	return &BaseAdapter{
		meta:       meta,
		httpClient: NewSafeHTTPClient(timeout, allowLocal),
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

	if b.meta.BaseURL != "" && (strings.HasPrefix(req.Path, "http://") || strings.HasPrefix(req.Path, "https://")) {
		baseU, err := url.Parse(b.meta.BaseURL)
		targetU, err2 := url.Parse(targetURL)
		if err == nil && err2 == nil && baseU.Host != "" && !strings.EqualFold(targetU.Host, baseU.Host) {
			return nil, fmt.Errorf("target host %q does not match provider base host %q", targetU.Host, baseU.Host)
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

func NewSafeHTTPClient(timeout time.Duration, allowLocal bool) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if !allowLocal {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ip := net.ParseIP(host)
				if ip != nil {
					if isPrivateIP(ip) {
						return nil, fmt.Errorf("SSRF protection: target IP %s is not permitted", ip.String())
					}
					return dialer.DialContext(ctx, network, addr)
				}
				ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
				if err != nil {
					return nil, fmt.Errorf("SSRF DNS resolution failed: %w", err)
				}
				if len(ips) == 0 {
					return nil, fmt.Errorf("SSRF DNS resolution returned no IPs for %s", host)
				}
				for _, rip := range ips {
					if isPrivateIP(rip) {
						return nil, fmt.Errorf("SSRF protection: host %s resolved to private IP %s", host, rip.String())
					}
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
			}
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			if !allowLocal {
				if err := ValidateSSRF(req.URL.String()); err != nil {
					return fmt.Errorf("redirect blocked by SSRF protection: %w", err)
				}
			}
			return nil
		},
	}
}

func ValidateSSRF(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("forbidden protocol: %s (only http and https allowed)", scheme)
	}

	hostname := strings.ToLower(u.Hostname())
	hostname = strings.TrimSuffix(hostname, ".")
	if hostname == "" {
		return errors.New("empty hostname")
	}

	if hostname == "localhost" || hostname == "0.0.0.0" || hostname == "127.0.0.1" || hostname == "::1" ||
		hostname == "169.254.169.254" || hostname == "100.100.100.200" || hostname == "168.63.129.16" ||
		hostname == "metadata.google.internal" || hostname == "metadata.internal" {
		return fmt.Errorf("target host %s is not permitted (loopback/metadata)", hostname)
	}

	if strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".internal") ||
		strings.HasSuffix(hostname, ".localhost") || strings.HasSuffix(hostname, ".arpa") {
		return fmt.Errorf("target internal domain %s is not permitted", hostname)
	}

	ip := net.ParseIP(hostname)
	if ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("target IP %s is within private or metadata range", ip.String())
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", hostname)
	if err == nil {
		for _, resolvedIP := range ips {
			if isPrivateIP(resolvedIP) {
				return fmt.Errorf("target host %s resolved to private/metadata IP %s", hostname, resolvedIP.String())
			}
		}
	}

	return nil
}

func IsPrivateIP(ip net.IP) bool {
	return isPrivateIP(ip)
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	ipv4 := ip.To4()
	if ipv4 != nil {
		switch {
		case ipv4[0] == 0:
			return true
		case ipv4[0] == 10:
			return true
		case ipv4[0] == 100 && ipv4[1] >= 64 && ipv4[1] <= 127:
			return true
		case ipv4[0] == 127:
			return true
		case ipv4[0] == 168 && ipv4[1] == 63 && ipv4[2] == 129 && ipv4[3] == 16:
			return true
		case ipv4[0] == 169 && ipv4[1] == 254:
			return true
		case ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31:
			return true
		case ipv4[0] == 192 && ipv4[1] == 0 && ipv4[2] == 0:
			return true
		case ipv4[0] == 192 && ipv4[1] == 0 && ipv4[2] == 2:
			return true
		case ipv4[0] == 192 && ipv4[1] == 168:
			return true
		case ipv4[0] == 198 && ipv4[1] >= 18 && ipv4[1] <= 19:
			return true
		case ipv4[0] == 198 && ipv4[1] == 51 && ipv4[2] == 100:
			return true
		case ipv4[0] == 203 && ipv4[1] == 0 && ipv4[2] == 113:
			return true
		case ipv4[0] >= 224 && ipv4[0] <= 239:
			return true
		case ipv4[0] >= 240:
			return true
		case ipv4[0] == 100 && ipv4[1] == 100 && ipv4[2] == 100 && ipv4[3] == 200:
			return true
		case ipv4[0] == 255 && ipv4[1] == 255 && ipv4[2] == 255 && ipv4[3] == 255:
			return true
		}
		return false
	}

	if len(ip) == 16 {
		if ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] == 0 &&
			ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
			ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0 {
			embedded := net.IPv4(ip[12], ip[13], ip[14], ip[15])
			return isPrivateIP(embedded)
		}

		if ip[0] == 0x20 && ip[1] == 0x02 {
			embedded := net.IPv4(ip[2], ip[3], ip[4], ip[5])
			return isPrivateIP(embedded)
		}

		if ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b &&
			ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
			ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0 {
			embedded := net.IPv4(ip[12], ip[13], ip[14], ip[15])
			return isPrivateIP(embedded)
		}

		if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x00 && ip[3] == 0x00 {
			embedded := net.IPv4(^ip[12], ^ip[13], ^ip[14], ^ip[15])
			return isPrivateIP(embedded)
		}

		if (ip[0] & 0xfe) == 0xfc {
			return true
		}
		if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return true
		}
		if ip[0] == 0xff {
			return true
		}
		if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8 {
			return true
		}
		if ip[0] == 0x01 && ip[1] == 0x00 {
			return true
		}
		if ip[0] == 0x20 && ip[1] == 0x01 && (ip[2] == 0x10 || ip[2] == 0x20) {
			return true
		}
	}

	return false
}
