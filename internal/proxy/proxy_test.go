package proxy

import (
	"net"
	"testing"
	"time"
)

func TestProfileURL(t *testing.T) {
	p := &Profile{
		Scheme:   "http",
		Host:     "proxy.example.com",
		Port:     8080,
		Username: "user",
		Password: "password",
	}
	u, err := p.URL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.String() != "http://user:password@proxy.example.com:8080" {
		t.Fatalf("expected http://user:password@proxy.example.com:8080, got %s", u.String())
	}
}

func TestSSRFBlocked(t *testing.T) {
	m := NewManager(false)

	blockedHosts := []string{"127.0.0.1", "10.1.2.3", "192.168.1.1", "172.16.0.5"}
	for _, h := range blockedHosts {
		if err := m.ValidateDestination(h); err == nil {
			t.Errorf("expected SSRF error for %s", h)
		}
	}
}

func TestSSRFAllowedWhenEnabled(t *testing.T) {
	m := NewManager(true)

	if err := m.ValidateDestination("127.0.0.1"); err != nil {
		t.Errorf("expected local allowed when configured, got error: %v", err)
	}
}

func TestIsPrivateOrLocal(t *testing.T) {
	if !isPrivateOrLocal(net.ParseIP("127.0.0.1")) {
		t.Error("expected 127.0.0.1 to be private/local")
	}
	if !isPrivateOrLocal(net.ParseIP("10.0.0.1")) {
		t.Error("expected 10.0.0.1 to be private/local")
	}
	if !isPrivateOrLocal(net.ParseIP("192.168.0.1")) {
		t.Error("expected 192.168.0.1 to be private/local")
	}
	if !isPrivateOrLocal(net.ParseIP("172.20.0.1")) {
		t.Error("expected 172.20.0.1 to be private/local")
	}
	if isPrivateOrLocal(net.ParseIP("8.8.8.8")) {
		t.Error("expected 8.8.8.8 to be public")
	}
}

func TestManagerTransportCaching(t *testing.T) {
	m := NewManager(false)

	client1, err := m.GetClient(nil, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}

	client2, err := m.GetClient(nil, 10*time.Second)
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}

	if client1.Transport != client2.Transport {
		t.Error("expected reused shared transport for direct connections")
	}
}

func TestSocks5ProfileURL(t *testing.T) {
	p := &Profile{
		Scheme: "socks5",
		Host:   "103.253.213.185",
		Port:   1080,
	}
	u, err := p.URL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.String() != "socks5://103.253.213.185:1080" {
		t.Fatalf("expected socks5://103.253.213.185:1080, got %s", u.String())
	}
}

func TestDirectProfileTransport(t *testing.T) {
	m := NewManager(false)
	trDirect, err := m.GetTransport(&Profile{Scheme: "direct"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trDirect.Proxy != nil {
		t.Error("expected nil Proxy for direct profile")
	}
}

func TestRelayTransportAndEdgeDetection(t *testing.T) {
	pRelay := &Profile{
		Host:   "xiomi16799.xiomi16799.workers.dev",
		Port:   443,
		Scheme: "https",
	}
	if !pRelay.IsEdgeRelay() {
		t.Error("expected workers.dev to be edge relay")
	}
	if pRelay.RelayURL() != "https://xiomi16799.xiomi16799.workers.dev" {
		t.Errorf("unexpected relay url: %s", pRelay.RelayURL())
	}

	pVercel := &Profile{
		Host:   "app.vercel.app",
		Port:   80,
		Scheme: "http",
	}
	if !pVercel.IsEdgeRelay() {
		t.Error("expected vercel.app to be edge relay")
	}

	pDeno := &Profile{
		Host: "test.deno.net",
	}
	if !pDeno.IsEdgeRelay() {
		t.Error("expected deno.net to be edge relay")
	}

	pDirect := &Profile{
		Host:   "my-proxy.com",
		Port:   8080,
		Scheme: "http",
	}
	if pDirect.IsEdgeRelay() {
		t.Error("standard proxy should not be edge relay")
	}
}
