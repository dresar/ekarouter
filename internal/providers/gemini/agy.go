package gemini

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/dresar/ekarouter/internal/providers"
)

var toolNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_.:\-]`)

type ThoughtSignatureStore struct {
	mu         sync.RWMutex
	signatures map[string]string
}

func NewThoughtSignatureStore() *ThoughtSignatureStore {
	return &ThoughtSignatureStore{
		signatures: make(map[string]string),
	}
}

func (s *ThoughtSignatureStore) Store(sessionID, signature string) {
	if sessionID == "" || signature == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.signatures[sessionID] = signature
}

func (s *ThoughtSignatureStore) Get(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.signatures[sessionID]
}

func (s *ThoughtSignatureStore) Clear(sessionID string) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.signatures, sessionID)
}

func SanitizeToolName(name string) string {
	if name == "" {
		return "tool"
	}
	cleaned := toolNameRegex.ReplaceAllString(name, "_")
	if !((cleaned[0] >= 'a' && cleaned[0] <= 'z') || (cleaned[0] >= 'A' && cleaned[0] <= 'Z') || cleaned[0] == '_') {
		cleaned = "tool_" + cleaned
	}
	if len(cleaned) > 64 {
		cleaned = cleaned[:64]
	}
	return cleaned
}

func ExtractAspectRatio(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, ar := range []string{"16:9", "9:16", "4:3", "3:4", "1:1"} {
		if strings.Contains(lower, ar) {
			return ar
		}
	}
	return "1:1"
}

func BuildAntigravityHeaders(creds *providers.Credentials) http.Header {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	h.Set("User-Agent", "Antigravity-IDE/1.0")
	h.Set("X-Goog-Api-Client", "gl-go/ gccl/ antigravity")
	if creds != nil {
		if creds.AccessToken != "" {
			h.Set("Authorization", "Bearer "+creds.AccessToken)
		} else if creds.APIKey != "" {
			h.Set("Authorization", "Bearer "+creds.APIKey)
		}
	}
	return h
}

func BuildAntigravityURL(baseURL string) string {
	cleaned := strings.TrimRight(baseURL, "/")
	if cleaned == "" {
		return "https://daily-cloudcode-pa.googleapis.com/v1internal:generateContent"
	}
	if strings.Contains(cleaned, ":generateContent") {
		return cleaned
	}
	return cleaned + "/v1internal:generateContent"
}

func WrapAntigravityRequest(projectID, model string, body map[string]any) map[string]any {
	return map[string]any{
		"projectId": projectID,
		"model":     model,
		"request":   body,
	}
}
