package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// Config stores encryption configuration
type Config struct {
	Algorithm    string `json:"algorithm"`
	KeyDerivation struct {
		Algorithm   string `json:"algorithm"`
		Iterations  int    `json:"iterations"`
		SaltLength  int    `json:"salt_length"`
	} `json:"key_derivation"`
	IVLength  int    `json:"iv_length"`
	TagLength int    `json:"tag_length"`
	KeyLength int    `json:"key_length"`
	Salt      string `json:"salt"`
	Key       string `json:"key"`
}

// Service provides end-to-end encryption functionality
type Service struct {
	config Config
	key    []byte
}

// NewService creates a new encryption service
func NewService(configPath string) (*Service, error) {
	var config Config
	
	// Load config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	
	// Decode the base64 key
	key, err := base64.StdEncoding.DecodeString(config.Key)
	if err != nil {
		return nil, err
	}
	
	// If key length doesn't match, derive it using PBKDF2
	if len(key) != config.KeyLength {
		salt, err := base64.StdEncoding.DecodeString(config.Salt)
		if err != nil {
			return nil, err
		}
		
		key = pbkdf2.Key(key, salt, config.KeyDerivation.Iterations, config.KeyLength, sha256.New)
	}
	
	return &Service{
		config: config,
		key:    key,
	}, nil
}

// Encrypt encrypts data using AES-GCM
func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	// Create a new cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	// Encrypt and seal
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	// Create a new cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Check if the ciphertext is valid
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	
	// Split nonce and ciphertext
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// EncryptString encrypts a string and returns a base64-encoded string
func (s *Service) EncryptString(plaintext string) (string, error) {
	encrypted, err := s.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64-encoded string
func (s *Service) DecryptString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	
	decrypted, err := s.Decrypt(data)
	if err != nil {
		return "", err
	}
	
	return string(decrypted), nil
}

// GenerateKeyPair generates a new key pair for asymmetric encryption
func (s *Service) GenerateKeyPair() (publicKey, privateKey string, err error) {
	// This is a simplified version for demonstration
	// In a real implementation, use RSA or ECC
	publicKeyBytes := make([]byte, 32)
	privateKeyBytes := make([]byte, 32)
	
	_, err = rand.Read(publicKeyBytes)
	if err != nil {
		return "", "", err
	}
	
	_, err = rand.Read(privateKeyBytes)
	if err != nil {
		return "", "", err
	}
	
	publicKey = base64.StdEncoding.EncodeToString(publicKeyBytes)
	privateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)
	
	return publicKey, privateKey, nil
} 