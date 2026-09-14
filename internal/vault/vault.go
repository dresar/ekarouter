package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type CredentialType string

const (
	TypeAPIKey         CredentialType = "api_key"
	TypeBearerToken    CredentialType = "bearer_token"
	TypePersonalToken  CredentialType = "personal_access_token"
	TypeOAuthToken     CredentialType = "oauth_token"
	TypeOAuthRefresh   CredentialType = "oauth_refresh_token"
	TypeServiceAccount CredentialType = "service_account"
	TypeCloudAccessKey CredentialType = "cloud_access_key"
	TypeBasicAuth      CredentialType = "basic_auth"
	TypeWebhookSecret  CredentialType = "webhook_secret"
	TypeSigningSecret  CredentialType = "signing_secret"
	TypeCustomSecret   CredentialType = "custom_secret"
)

type HealthState string

const (
	HealthUnknown   HealthState = "unknown"
	HealthHealthy   HealthState = "healthy"
	HealthDegraded  HealthState = "degraded"
	HealthUnhealthy HealthState = "unhealthy"
	HealthExpired   HealthState = "expired"
)

type Credential struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	CredentialType CredentialType `json:"credential_type"`
	ProviderID     string         `json:"provider_id"`
	ProjectID      string         `json:"project_id,omitempty"`
	TeamID         string         `json:"team_id,omitempty"`
	Environment    string         `json:"environment"`
	EncryptedValue string         `json:"-"`
	MaskedValue    string         `json:"masked_value"`
	Priority       int            `json:"priority"`
	Tags           string         `json:"tags"`
	Status         string         `json:"status"`
	HealthState    HealthState    `json:"health_state"`
	CooldownUntil  *time.Time     `json:"cooldown_until,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	Notes          string         `json:"notes"`
	LastUsedAt     *time.Time     `json:"last_used_at,omitempty"`
	LastValidated  *time.Time     `json:"last_validated_at,omitempty"`
	LastError      string         `json:"last_error,omitempty"`
	RequestCount   int64          `json:"request_count"`
	ErrorCount     int64          `json:"error_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Vault struct {
	key []byte
}

func NewVault(secretKey string) (*Vault, error) {
	if len(secretKey) < 16 {
		return nil, errors.New("vault secret key must be at least 16 characters")
	}
	hash := sha256.Sum256([]byte(secretKey))
	return &Vault{key: hash[:]}, nil
}

func (v *Vault) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("v1:%s", base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func (v *Vault) Decrypt(payload string) (string, error) {
	if payload == "" {
		return "", nil
	}
	dataStr := payload
	if strings.HasPrefix(payload, "v1:") {
		dataStr = strings.TrimPrefix(payload, "v1:")
	}
	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("malformed vault ciphertext")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func MaskCredential(raw string) string {
	val := strings.TrimSpace(raw)
	if val == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(val), "bearer ") {
		token := strings.TrimSpace(val[7:])
		if len(token) >= 4 {
			return "Bearer ****" + token[len(token)-4:]
		}
		return "Bearer ****"
	}
	if strings.HasPrefix(val, "basic ") {
		token := strings.TrimSpace(val[6:])
		if len(token) >= 4 {
			return "Basic ****" + token[len(token)-4:]
		}
		return "Basic ****"
	}
	return maskToken(val)
}

func maskToken(token string) string {
	l := len(token)
	if l <= 4 {
		return "****"
	}
	if l <= 8 {
		return token[:2] + "****" + token[l-2:]
	}

	idx := strings.IndexAny(token, "_-")
	if idx > 0 && idx < l-4 && idx <= 6 {
		prefix := token[:idx+1]
		suffix := token[l-4:]
		return prefix + "****" + suffix
	}

	if l > 12 {
		return token[:4] + "****" + token[l-4:]
	}
	return token[:2] + "****" + token[l-2:]
}
