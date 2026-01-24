package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "hostodo-cli"
	accountName = "access-token"
)

// TokenStore manages CLI token storage
type TokenStore struct {
	fallbackPath string
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	home, _ := os.UserHomeDir()
	return &TokenStore{
		fallbackPath: filepath.Join(home, ".hostodo", "token.enc"),
	}
}

// Save stores a token in keychain, falling back to encrypted file
func (s *TokenStore) Save(token string) error {
	err := keyring.Set(serviceName, accountName, token)
	if err == nil {
		// Also delete any fallback file if keychain succeeds
		os.Remove(s.fallbackPath)
		return nil
	}

	// Fallback to encrypted file
	fmt.Println("Warning: System keychain unavailable, using encrypted file storage")
	return s.saveToFile(token)
}

// Get retrieves token from keychain or fallback file
func (s *TokenStore) Get() (string, error) {
	token, err := keyring.Get(serviceName, accountName)
	if err == nil {
		return token, nil
	}

	// Try fallback file
	return s.getFromFile()
}

// Delete removes token from keychain and fallback file
func (s *TokenStore) Delete() error {
	// Delete from keychain (ignore error if not found)
	keyring.Delete(serviceName, accountName)
	// Delete fallback file (ignore error if not found)
	os.Remove(s.fallbackPath)
	return nil
}

// saveToFile encrypts and saves token to file
func (s *TokenStore) saveToFile(token string) error {
	key := s.deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.fallbackPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write with restrictive permissions
	if err := os.WriteFile(s.fallbackPath, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// getFromFile decrypts and returns token from file
func (s *TokenStore) getFromFile() (string, error) {
	data, err := os.ReadFile(s.fallbackPath)
	if err != nil {
		return "", fmt.Errorf("not authenticated: %w", err)
	}

	key := s.deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("invalid token file")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return string(plaintext), nil
}

// deriveKey generates encryption key from machine-specific data
func (s *TokenStore) deriveKey() []byte {
	hostname, _ := os.Hostname()
	hash := sha256.Sum256([]byte("hostodo-cli-" + hostname))
	return hash[:]
}

// Helper functions for package-level access
var defaultStore = NewTokenStore()

// GetToken retrieves the stored access token
func GetToken() (string, error) {
	return defaultStore.Get()
}

// SaveToken stores an access token
func SaveToken(token string) error {
	return defaultStore.Save(token)
}

// DeleteToken removes the stored token
func DeleteToken() error {
	return defaultStore.Delete()
}

// IsAuthenticated checks if a token exists
func IsAuthenticated() bool {
	token, err := GetToken()
	return err == nil && token != ""
}
