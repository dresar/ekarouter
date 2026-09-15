package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbeddingsHandler_SingleString(t *testing.T) {
	h := &GatewayHandler{}

	body := `{"model":"text-embedding-3-small","input":"The quick brown fox jumps over the lazy dog"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Embeddings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp EmbeddingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %s", resp.Object)
	}
	if resp.Model != "text-embedding-3-small" {
		t.Errorf("expected model 'text-embedding-3-small', got %s", resp.Model)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(resp.Data))
	}
	if resp.Data[0].Index != 0 || resp.Data[0].Object != "embedding" {
		t.Errorf("invalid embedding metadata: %+v", resp.Data[0])
	}
	vec, ok := resp.Data[0].Embedding.([]any)
	if !ok || len(vec) != 1536 {
		t.Errorf("expected 1536 float elements, got %T with len %d", resp.Data[0].Embedding, len(vec))
	}
	if resp.Usage.TotalTokens <= 0 {
		t.Errorf("expected positive total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestEmbeddingsHandler_ArrayInput(t *testing.T) {
	h := &GatewayHandler{}

	body := `{"model":"text-embedding-3-small","input":["First document chunk","Second document chunk","Third document chunk"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Embeddings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp EmbeddingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 embeddings, got %d", len(resp.Data))
	}
	for i := 0; i < 3; i++ {
		if resp.Data[i].Index != i {
			t.Errorf("expected index %d, got %d", i, resp.Data[i].Index)
		}
	}
}

func TestEmbeddingsHandler_CustomDimensions(t *testing.T) {
	h := &GatewayHandler{}

	body := `{"model":"text-embedding-3-large","input":"custom dimensions test","dimensions":512}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Embeddings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp EmbeddingsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	vec, ok := resp.Data[0].Embedding.([]any)
	if !ok || len(vec) != 512 {
		t.Errorf("expected 512 dimensions, got len %d", len(vec))
	}
}

func TestEmbeddingsHandler_Base64Encoding(t *testing.T) {
	h := &GatewayHandler{}

	body := `{"model":"text-embedding-3-small","input":"base64 encoding format","encoding_format":"base64","dimensions":128}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Embeddings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp EmbeddingsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	b64Str, ok := resp.Data[0].Embedding.(string)
	if !ok {
		t.Fatalf("expected string base64 embedding, got %T", resp.Data[0].Embedding)
	}
	decoded, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		t.Fatalf("failed to decode base64 embedding: %v", err)
	}
	if len(decoded) != 128*4 {
		t.Errorf("expected %d bytes for 128 float32 values, got %d", 128*4, len(decoded))
	}
}

func TestEmbeddingsHandler_ValidationErrors(t *testing.T) {
	h := &GatewayHandler{}

	tests := []struct {
		name string
		body string
	}{
		{"missing model", `{"input":"test"}`},
		{"empty model", `{"model":" ","input":"test"}`},
		{"missing input", `{"model":"text-embedding-3-small"}`},
		{"empty input string", `{"model":"text-embedding-3-small","input":""}`},
		{"empty input array", `{"model":"text-embedding-3-small","input":[]}`},
		{"array with empty string", `{"model":"text-embedding-3-small","input":["valid",""]}`},
		{"numeric input", `{"model":"text-embedding-3-small","input":12345}`},
		{"invalid json", `{"model": invalid json}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.Embeddings(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 Bad Request for %s, got %d", tt.name, rec.Code)
			}
		})
	}
}
