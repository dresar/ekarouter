package httpapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	apiKeyIDKey  contextKey = "api_key_id"
	userIDKey    contextKey = "user_id"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 12)
			_, _ = rand.Read(b)
			reqID = "req_" + hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BodyLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	originsMap := make(map[string]bool)
	allowAll := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		originsMap[strings.ToLower(o)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll || originsMap[strings.ToLower(origin)] {
				if allowAll && origin == "" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, Accept, X-Token-Saver")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GatewayAuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			tokenHash := auth.HashToken(token)

			var keyID string
			err := db.QueryRowContext(r.Context(), "SELECT id FROM api_keys WHERE hash = ? AND enabled = 1", tokenHash).Scan(&keyID)
			if err != nil {
				http.Error(w, `{"error":"unauthorized: invalid api key"}`, http.StatusUnauthorized)
				return
			}

			go func(k string) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_, _ = db.ExecContext(ctx, "UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?", k)
			}(keyID)

			ctx := context.WithValue(r.Context(), apiKeyIDKey, keyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SessionAuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			if cookie, err := r.Cookie("session_token"); err == nil {
				token = cookie.Value
			}
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if token == "" {
				http.Error(w, `{"error":"unauthorized session"}`, http.StatusUnauthorized)
				return
			}

			tokenHash := auth.HashToken(token)
			var userID string
			var expiresAt time.Time
			var revokedAt sql.NullTime

			err := db.QueryRowContext(r.Context(),
				"SELECT user_id, expires_at, revoked_at FROM sessions WHERE token_hash = ?",
				tokenHash).Scan(&userID, &expiresAt, &revokedAt)

			if err != nil || revokedAt.Valid || time.Now().After(expiresAt) {
				http.Error(w, `{"error":"unauthorized or expired session"}`, http.StatusUnauthorized)
				return
			}

			_, _ = db.ExecContext(r.Context(), "UPDATE sessions SET last_seen_at = CURRENT_TIMESTAMP WHERE token_hash = ?", tokenHash)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}
