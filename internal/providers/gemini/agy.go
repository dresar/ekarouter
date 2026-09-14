package gemini

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
)

const DefaultThinkingAGSignature = "EuwGCukGAXLI2nxwZIq54WWSoL/YN0P3TsDZ7zRnLi8g0S4aVr2HUGxvaHKySuY6HAVzcE0GPGjXrytLIldxthSvfxgUlJh6Qa9Z+Oj5QZBlYdg6HaJ6yuY5R7waE6rdwBsRf7Ft2j3DJ9rMi9qhWFqApewYtPhls3VHtuvND3l8Rm09+lbAXQs6KKWEWrxNLKTBkfpMgXhRERc/TQRMZu1twAablm6/Zk1tsYRvfWKLsNbeKF+CCojJdXJKvnR/8Ouuoa+Y2Ti20hcW7aZIIjZDFYPU//k6Ybmhg69J/imbFai2ckhfLaisqdDkdoIiBJScTOUvYqP6AE9d4MsydSC+UlhIMk4hoP76R8vUSCZRMkjOaDXstf/QoVZKbt94wyRZgAJ1G0BqI8L5ow86kLpA4wJEtxsRGymOE4bKUvApveBakYDNM9APkf+LbtbzWSseGjoZcSlycF9iN8Q2XNYKRrHbv3Lr5Y8JjdH/5y/6SHkNehTEZugaeGnSPSyCTWto1kQgHpxdWmhkLfJGNUGLmue7Mesj4TSms4J33mRpYVhNB/J333FCqIP0hr/E7BkkjEn7yZ4X7SQlh+xKPurapsnHRwiKmtsilmEFrnTE9iQr+pMr6M29qqFNv1tr5yumbaJw8JW9sB15tNsRv+dW6BjNanbsKz7HCgKUBc8tGy+7YuhXzAfViyRefcjK7eZW0Fbyt7AbybJTKz78W8NH7ye6LAwzOebXpeZ4D43fNIt8bKh26qgduSQv/7o+pAflkuqHZ99YWgHQ8h8OkZFi3eOiSYjsjhdZ/czWOdoPI/OnqIldzMPF5YlrKBLFX8VhRKVmqgsmWf5PHGulHhMkVlS+XG2UIseGy69ARa93D78Gsa+1n1kJr7EEB7Rh+27vUMxVYLdz1yMSvE5nalTAlg/ZeG8+XQ0cHuAI3KbQpHW2Q++RdXfm5JzD5WdJZUU+Zn8t8UUn85BH4RxZLeE0qJikgSsKoYVBc6YhiMjhPgkR95ReimY4Z0xCJdRo1gjexOFeODZMpQF6Yxnoic7IrdgsFA3iePTbFnPp3IAM1fAThWhXJUn3QInUOTd5o1qmTmn6REbL15g/JQNl+dqUoPkhleeb2V3kjqp1okmO3wMZbPknR3S1LZNmlS72/iBQUm+n2b/RCn4PjmM2"

var toolNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_.:\-]`)

type signatureEntry struct {
	signature string
	expiresAt time.Time
}

type ThoughtSignatureStore struct {
	mu         sync.RWMutex
	signatures map[string]signatureEntry
	order      []string
	maxSize    int
	ttl        time.Duration
}

func NewThoughtSignatureStore() *ThoughtSignatureStore {
	return &ThoughtSignatureStore{
		signatures: make(map[string]signatureEntry),
		order:      make([]string, 0, 1024),
		maxSize:    2000,
		ttl:        time.Hour,
	}
}

func (s *ThoughtSignatureStore) Store(key, signature string) {
	if key == "" || signature == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if len(s.signatures) >= s.maxSize {
		pruneIndex := 0
		for i, k := range s.order {
			if entry, ok := s.signatures[k]; ok {
				if now.After(entry.expiresAt) || i < len(s.order)/4 {
					delete(s.signatures, k)
					pruneIndex = i + 1
				}
			}
		}
		if pruneIndex > 0 {
			s.order = s.order[pruneIndex:]
		}
		if len(s.signatures) >= s.maxSize && len(s.order) > 0 {
			oldest := s.order[0]
			s.order = s.order[1:]
			delete(s.signatures, oldest)
		}
	}

	if _, exists := s.signatures[key]; !exists {
		s.order = append(s.order, key)
	}
	s.signatures[key] = signatureEntry{
		signature: signature,
		expiresAt: now.Add(s.ttl),
	}
}

func (s *ThoughtSignatureStore) Get(key string) string {
	if key == "" {
		return ""
	}
	s.mu.RLock()
	entry, ok := s.signatures[key]
	s.mu.RUnlock()
	if !ok {
		return ""
	}
	if time.Now().After(entry.expiresAt) {
		s.Clear(key)
		return ""
	}
	return entry.signature
}

func (s *ThoughtSignatureStore) Clear(key string) {
	if key == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.signatures, key)
}

func (s *ThoughtSignatureStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.signatures)
}

func IsImageModel(model string) bool {
	lower := strings.ToLower(model)
	return strings.Contains(lower, "image") || strings.HasPrefix(lower, "imagen")
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
		for k, v := range creds.Headers {
			h.Set(k, v)
		}
	}
	return h
}

func BuildAntigravityURL(baseURL string, stream bool, isImageModel bool) string {
	cleaned := strings.TrimRight(baseURL, "/")
	if cleaned == "" {
		cleaned = "https://daily-cloudcode-pa.googleapis.com"
	}
	action := "generateContent"
	if stream && !isImageModel {
		action = "streamGenerateContent?alt=sse"
	}
	if strings.Contains(cleaned, "/v1internal:") {
		idx := strings.Index(cleaned, "/v1internal:")
		cleaned = cleaned[:idx]
	}
	return fmt.Sprintf("%s/v1internal:%s", cleaned, action)
}

func WrapAntigravityRequest(projectID, model string, body map[string]any) map[string]any {
	return map[string]any{
		"projectId": projectID,
		"model":     model,
		"request":   body,
	}
}
