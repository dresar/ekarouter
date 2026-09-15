package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
)

type Profile struct {
	ID          string
	Name        string
	Scheme      string
	Host        string
	Port        int
	Username    string
	Password    string
	NoProxy     string
	StrictProxy bool
	RelayType   string
	RelayConfig string
}

func ParseProxyURL(rawURL string) (*Profile, error) {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	host := u.Hostname()
	portStr := u.Port()
	port := 80
	if u.Scheme == "https" || u.Scheme == "relay" {
		port = 443
	} else if u.Scheme == "socks5" {
		port = 1080
	}
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "http"
	}
	return &Profile{
		Scheme:   scheme,
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
	}, nil
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
		key = fmt.Sprintf("%s://%s:%s@%s:%d?no_proxy=%s", profile.Scheme, profile.Username, profile.Password, profile.Host, profile.Port, profile.NoProxy)
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
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
			port = "80"
		}
		if err := m.ValidateDestination(host); err != nil {
			return nil, err
		}
		if !m.allowLocalProviders {
			ip := net.ParseIP(host)
			if ip != nil {
				return dialer.DialContext(ctx, network, addr)
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("SSRF resolution failed for %s: %w", host, err)
			}
			targetAddr := net.JoinHostPort(ips[0].String(), port)
			return dialer.DialContext(ctx, network, targetAddr)
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
		if profile != nil && profile.NoProxy != "" {
			rawParts := strings.Split(profile.NoProxy, ",")
			var noProxyList []string
			for _, p := range rawParts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					noProxyList = append(noProxyList, strings.ToLower(trimmed))
				}
			}
			newTr.Proxy = func(req *http.Request) (*url.URL, error) {
				reqHost := strings.ToLower(req.URL.Hostname())
				for _, np := range noProxyList {
					if reqHost == np || strings.HasSuffix(reqHost, np) || strings.HasSuffix(reqHost, "."+np) {
						return nil, nil
					}
				}
				return proxyURL, nil
			}
		} else {
			newTr.Proxy = http.ProxyURL(proxyURL)
		}
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

	trimmedHost := strings.TrimSuffix(strings.ToLower(host), ".")
	if trimmedHost == "localhost" || trimmedHost == "0.0.0.0" || trimmedHost == "127.0.0.1" || trimmedHost == "::1" ||
		trimmedHost == "169.254.169.254" || trimmedHost == "100.100.100.200" || trimmedHost == "168.63.129.16" ||
		trimmedHost == "metadata.google.internal" || trimmedHost == "metadata.internal" {
		return errors.New("access to private, local, or loopback network blocked by SSRF policy")
	}

	if strings.HasSuffix(trimmedHost, ".local") || strings.HasSuffix(trimmedHost, ".internal") ||
		strings.HasSuffix(trimmedHost, ".localhost") || strings.HasSuffix(trimmedHost, ".arpa") {
		return errors.New("access to private, local, or loopback network blocked by SSRF policy")
	}

	ip := net.ParseIP(trimmedHost)
	if ip != nil {
		if platform.IsPrivateIP(ip) {
			return errors.New("access to private, local, or loopback network blocked by SSRF policy")
		}
		return nil
	}

	ips, err := net.LookupIP(trimmedHost)
	if err != nil {
		return fmt.Errorf("lookup host failed: %w", err)
	}
	if len(ips) == 0 {
		return errors.New("lookup host returned no IP addresses")
	}

	for _, rip := range ips {
		if platform.IsPrivateIP(rip) {
			return errors.New("access to private, local, or loopback network blocked by SSRF policy")
		}
	}

	return nil
}

func isPrivateOrLocal(ip net.IP) bool {
	return platform.IsPrivateIP(ip)
}
