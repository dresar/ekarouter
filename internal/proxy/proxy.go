package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Profile struct {
	ID       string
	Name     string
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
}

func (p *Profile) URL() (*url.URL, error) {
	if p == nil || p.Host == "" {
		return nil, nil
	}
	scheme := p.Scheme
	if scheme == "" {
		scheme = "http"
	}
	raw := fmt.Sprintf("%s://%s:%d", scheme, p.Host, p.Port)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if p.Username != "" {
		if p.Password != "" {
			u.User = url.UserPassword(p.Username, p.Password)
		} else {
			u.User = url.User(p.Username)
		}
	}
	return u, nil
}

func (p *Profile) IsEdgeRelay() bool {
	if p == nil {
		return false
	}
	h := strings.ToLower(p.Host)
	return strings.Contains(h, "workers.dev") || strings.Contains(h, "vercel.app") || strings.Contains(h, "deno.net") || p.Scheme == "relay"
}

func (p *Profile) RelayURL() string {
	if p == nil || p.Host == "" {
		return ""
	}
	scheme := p.Scheme
	if scheme == "" || scheme == "relay" {
		scheme = "https"
	}
	if p.Port == 443 || p.Port == 80 || p.Port == 0 {
		return fmt.Sprintf("%s://%s", scheme, p.Host)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, p.Host, p.Port)
}

type RelayTransport struct {
	RelayURL string
	Base     http.RoundTripper
}

func (t *RelayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	relayReq := req.Clone(req.Context())
	parsedRelay, err := url.Parse(t.RelayURL)
	if err != nil {
		return nil, err
	}

	relayReq.Header.Set("x-relay-target", fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host))
	relayReq.Header.Set("x-relay-path", req.URL.RequestURI())

	relayReq.URL.Scheme = parsedRelay.Scheme
	relayReq.URL.Host = parsedRelay.Host
	relayReq.URL.Path = parsedRelay.Path
	relayReq.URL.RawQuery = ""
	relayReq.Host = parsedRelay.Host

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(relayReq)
}

func TestProfile(ctx context.Context, p *Profile, timeout time.Duration) (bool, int, int64, error) {
	if p == nil {
		return false, 0, 0, errors.New("nil profile")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	start := time.Now()
	if p.IsEdgeRelay() {
		relayURL := p.RelayURL()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, relayURL, nil)
		if err != nil {
			return false, 0, time.Since(start).Milliseconds(), err
		}
		req.Header.Set("User-Agent", "EkaRouter/1.0")
		req.Header.Set("x-relay-target", "https://httpbin.org")
		req.Header.Set("x-relay-path", "/get")

		resp, err := client.Do(req)
		latency := time.Since(start).Milliseconds()
		if err != nil {
			return false, 0, latency, err
		}
		defer resp.Body.Close()
		ok := resp.StatusCode >= 200 && resp.StatusCode < 400
		return ok, resp.StatusCode, latency, nil
	}

	u, err := p.URL()
	if err != nil {
		return false, 0, 0, err
	}
	tr := &http.Transport{
		Proxy:           http.ProxyURL(u),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxyClient := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return false, 0, time.Since(start).Milliseconds(), err
	}
	resp, err := proxyClient.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0, latency, err
	}
	defer resp.Body.Close()
	ok := (resp.StatusCode >= 200 && resp.StatusCode < 400) || resp.StatusCode == 401
	return ok, resp.StatusCode, latency, nil
}

type Manager struct {
	mu                  sync.RWMutex
	transports          map[string]*http.Transport
	allowLocalProviders bool
}

func NewManager(allowLocalProviders bool) *Manager {
	return &Manager{
		transports:          make(map[string]*http.Transport),
		allowLocalProviders: allowLocalProviders,
	}
}

func (m *Manager) GetClient(profile *Profile, timeout time.Duration) (*http.Client, error) {
	if profile != nil && profile.IsEdgeRelay() {
		baseTr, err := m.GetTransport(nil)
		if err != nil {
			return nil, err
		}
		return &http.Client{
			Transport: &RelayTransport{
				RelayURL: profile.RelayURL(),
				Base:     baseTr,
			},
			Timeout: timeout,
		}, nil
	}
	transport, err := m.GetTransport(profile)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

func (m *Manager) GetTransport(profile *Profile) (*http.Transport, error) {
	key := "env"
	var proxyURL *url.URL
	var err error

	if profile != nil && profile.Host != "" && profile.Scheme != "direct" {
		key = fmt.Sprintf("%s://%s:%s@%s:%d", profile.Scheme, profile.Username, profile.Password, profile.Host, profile.Port)
		proxyURL, err = profile.URL()
		if err != nil {
			return nil, err
		}
	} else if profile != nil && profile.Scheme == "direct" {
		key = "direct"
	}

	m.mu.RLock()
	tr, ok := m.transports[key]
	m.mu.RUnlock()
	if ok {
		return tr, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if tr, ok := m.transports[key]; ok {
		return tr, nil
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		if err := m.ValidateDestination(host); err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, addr)
	}

	newTr := &http.Transport{
		DialContext:           customDialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if proxyURL != nil {
		newTr.Proxy = http.ProxyURL(proxyURL)
	} else if key == "direct" {
		newTr.Proxy = nil
	} else {
		newTr.Proxy = http.ProxyFromEnvironment
	}

	m.transports[key] = newTr
	return newTr, nil
}

func (m *Manager) ValidateDestination(host string) error {
	if m.allowLocalProviders {
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("lookup host failed: %w", err)
		}
		ips = []net.IP{ip}
	}

	for _, ip := range ips {
		if isPrivateOrLocal(ip) {
			return errors.New("access to private, local, or loopback network blocked by SSRF policy")
		}
	}

	return nil
}

func isPrivateOrLocal(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}

	ip4 := ip.To4()
	if ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		return false
	}

	if len(ip) == net.IPv6len {
		if ip[0] == 0xfc || ip[0] == 0xfd {
			return true
		}
	}

	return false
}
