package auth

import (
	"strings"
	"testing"
)

func TestCryptoServiceEncryptDecrypt(t *testing.T) {
	crypto, err := NewCryptoService("my-super-secret-key-32-characters-long!")
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	secret := "sk-ant-api03-abcdef123456789"
	encrypted, err := crypto.Encrypt(secret)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}
	if encrypted == secret {
		t.Fatal("encrypted text must not match plaintext")
	}

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
	if decrypted != secret {
		t.Fatalf("expected %s, got %s", secret, decrypted)
	}
}

func TestGenerateApiKey(t *testing.T) {
	rawKey, prefix, hash, err := GenerateApiKey()
	if err != nil {
		t.Fatalf("failed to generate api key: %v", err)
	}
	if !strings.HasPrefix(rawKey, "eka_live_") {
		t.Errorf("expected eka_live_ prefix, got %s", rawKey)
	}
	if !strings.HasPrefix(rawKey, prefix) {
		t.Errorf("prefix %s does not match raw key %s", prefix, rawKey)
	}
	if HashToken(rawKey) != hash {
		t.Errorf("hash does not match hashed raw key")
	}
}

func TestGenerateSessionToken(t *testing.T) {
	rawToken, tokenHash, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("failed to generate session: %v", err)
	}
	if len(rawToken) != 64 {
		t.Errorf("expected 64 char hex string, got %d", len(rawToken))
	}
	if HashToken(rawToken) != tokenHash {
		t.Errorf("hash does not match token hash")
	}
}
