package devtools

import (
	"testing"
)

func TestValidateRemoteURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://api.example.com/openapi.json", false},
		{"http://api.example.com/spec.json", false},
		{"ftp://files.example.com/spec", true},
		{"file:///etc/passwd", true},
		{"gopher://evil.com", true},
		{"https://localhost/api", true},
		{"https://127.0.0.1/api", true},
		{"https://[::1]/api", true},
		{"https://169.254.169.254/metadata", true},
		{"https://metadata.google.internal/v1", true},
		{"https://0.0.0.0/api", true},
		{"", true},
		{"not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := ValidateRemoteURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRemoteURL(%q) err=%v, wantErr=%v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestParseOpenAPIDocument(t *testing.T) {
	doc := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"servers": [{"url": "https://api.test.com/v1"}],
		"paths": {
			"/users": {
				"get": {"summary": "List users"},
				"post": {"summary": "Create user"}
			},
			"/users/{id}": {
				"get": {"summary": "Get user"},
				"delete": {"summary": "Delete user"}
			}
		},
		"components": {
			"securitySchemes": {
				"BearerAuth": {"type": "http", "scheme": "bearer"},
				"ApiKeyAuth": {"type": "apiKey", "in": "header", "name": "X-API-Key"}
			}
		}
	}`

	result, err := ParseOpenAPIDocument([]byte(doc))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if result.ProviderTemplate.ProviderName != "Test API" {
		t.Errorf("provider name: %s", result.ProviderTemplate.ProviderName)
	}
	if result.ProviderTemplate.BaseURL != "https://api.test.com/v1" {
		t.Errorf("base url: %s", result.ProviderTemplate.BaseURL)
	}

	if result.TotalCount != 4 {
		t.Errorf("expected 4 operations, got %d", result.TotalCount)
	}

	if result.DestructiveCount != 2 {
		t.Errorf("expected 2 destructive (POST + DELETE), got %d", result.DestructiveCount)
	}

	if len(result.AuthSchemes) != 2 {
		t.Errorf("expected 2 auth schemes, got %d", len(result.AuthSchemes))
	}

	for _, op := range result.Operations {
		if !op.Confirmed {
			continue
		}
		t.Error("operations should not be auto-confirmed")
	}
}

func TestParseInvalidDocument(t *testing.T) {
	_, err := ParseOpenAPIDocument([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDestructiveMethodDetection(t *testing.T) {
	doc := `{
		"openapi": "3.0.0",
		"info": {"title": "Destructive Test"},
		"paths": {
			"/data": {
				"get": {"summary": "safe read"},
				"post": {"summary": "create data"},
				"put": {"summary": "update data"},
				"delete": {"summary": "delete data"},
				"patch": {"summary": "patch data"}
			}
		}
	}`

	result, err := ParseOpenAPIDocument([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}

	destructive := 0
	for _, op := range result.Operations {
		if op.IsDestructive {
			destructive++
		}
	}
	if destructive != 4 {
		t.Errorf("expected 4 destructive ops (POST/PUT/DELETE/PATCH), got %d", destructive)
	}
}

func TestGenerateCurl(t *testing.T) {
	rt := &RequestTemplate{
		Method:        "POST",
		Path:          "/v1/chat/completions",
		Headers:       "Content-Type: application/json",
		CredentialRef: "EKA_OPENAI_KEY",
		BodySchema:    `{"model":"gpt-4","messages":[]}`,
		TimeoutMs:     30000,
	}

	curl := GenerateCurl(rt, "https://api.openai.com")
	if curl == "" {
		t.Fatal("empty curl output")
	}

	mustContain := []string{"curl", "-X", "POST", "api.openai.com/v1/chat/completions", "EKA_OPENAI_KEY"}
	for _, s := range mustContain {
		found := false
		for i := 0; i <= len(curl)-len(s); i++ {
			if curl[i:i+len(s)] == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("curl should contain %q, got:\n%s", s, curl)
		}
	}
}

func TestIsPrivateIP(t *testing.T) {
	privateIPs := []string{"10.0.0.1", "172.16.0.1", "192.168.1.1", "127.0.0.1", "169.254.1.1", "::1"}
	publicIPs := []string{"8.8.8.8", "1.1.1.1", "93.184.216.34"}

	for _, ip := range privateIPs {
		if !isPrivateIP(ip) {
			t.Errorf("%s should be private", ip)
		}
	}
	for _, ip := range publicIPs {
		if isPrivateIP(ip) {
			t.Errorf("%s should be public", ip)
		}
	}
}
