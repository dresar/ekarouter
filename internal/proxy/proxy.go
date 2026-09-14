package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
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
	raw := fmt.Sprintf("%s://%s:%d", p.Scheme, p.Host, p.Port)
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
	key := "direct"
	var proxyURL *url.URL
	var err error

	if profile != nil && profile.Host != "" {
		key = fmt.Sprintf("%s://%s:%s@%s:%d", profile.Scheme, profile.Username, profile.Password, profile.Host, profile.Port)
		proxyURL, err = profile.URL()
		if err != nil {
			return nil, err
		}
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
