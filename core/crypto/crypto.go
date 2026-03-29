package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"sync"

	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

// Argon2Params holds the parameters for argon2id hashing.
type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen int
}

// DefaultArgon2Params returns secure defaults for argon2id.
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}
}

// HashPassword hashes a password using argon2id.
func HashPassword(password string, params Argon2Params) (string, error) {
	salt := make([]byte, params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	// Format: $argon2id$v=19$m=MEMORY,t=TIME,p=THREADS$SALT$HASH
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory, params.Time, params.Threads, b64Salt, b64Hash), nil
}

// VerifyPassword verifies a password against an argon2id or bcrypt hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") || strings.HasPrefix(encodedHash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return err == nil, err
	}

	params, salt, hash, err := decodeArgon2Hash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

func decodeArgon2Hash(encodedHash string) (params Argon2Params, salt, hash []byte, err error) {
	var version int
	var memory uint32
	var time uint32
	var threads uint8
	var b64Salt, b64Hash string

	_, err = fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s",
		&version, &memory, &time, &threads, &b64Salt)
	if err != nil {
		// Try parsing the combined salt$hash part
		var rest string
		_, err = fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s",
			&version, &memory, &time, &threads, &rest)
		if err != nil {
			return params, nil, nil, fmt.Errorf("invalid argon2id hash format: %w", err)
		}
		parts := splitLast(rest, '$')
		if len(parts) != 2 {
			return params, nil, nil, fmt.Errorf("invalid argon2id hash: missing salt or hash")
		}
		b64Salt = parts[0]
		b64Hash = parts[1]
	}

	if b64Hash == "" {
		parts := splitLast(b64Salt, '$')
		if len(parts) == 2 {
			b64Salt = parts[0]
			b64Hash = parts[1]
		}
	}

	salt, err = base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return params, nil, nil, fmt.Errorf("decoding salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(b64Hash)
	if err != nil {
		return params, nil, nil, fmt.Errorf("decoding hash: %w", err)
	}

	params = Argon2Params{
		Time:    time,
		Memory:  memory,
		Threads: threads,
		KeyLen:  uint32(len(hash)),
		SaltLen: len(salt),
	}

	return params, salt, hash, nil
}

func splitLast(s string, sep byte) []string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// --- AES-256-GCM Encryption ---

// Encrypt encrypts plaintext using AES-256-GCM with the given key.
func Encrypt(plaintext []byte, keyHex string) ([]byte, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (64 hex chars), got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given key.
func Decrypt(ciphertext []byte, keyHex string) ([]byte, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// --- RSA/ECDSA Key Management ---

// KeyPair holds a signing key pair with metadata.
type KeyPair struct {
	ID         string
	Algorithm  string
	PrivateKey interface{}
	PublicKey  interface{}
}

// GenerateRSAKeyPair generates an RSA key pair for JWT signing.
func GenerateRSAKeyPair(bits int) (*KeyPair, error) {
	if bits < 2048 {
		bits = 2048
	}
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key: %w", err)
	}

	kid, err := GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		ID:         kid,
		Algorithm:  "RS256",
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}, nil
}

// GenerateECDSAKeyPair generates an ECDSA key pair for JWT signing.
func GenerateECDSAKeyPair() (*KeyPair, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating ECDSA key: %w", err)
	}

	kid, err := GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		ID:         kid,
		Algorithm:  "ES256",
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}, nil
}

// LoadRSAPrivateKeyFromFile loads an RSA private key from a PEM file.
func LoadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}
	return ParseRSAPrivateKey(data)
}

// ParseRSAPrivateKey parses an RSA private key from PEM data.
func ParseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return rsaKey, nil
}

// LoadRSAPublicKeyFromFile loads an RSA public key from a PEM file.
func LoadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}
	return ParseRSAPublicKey(data)
}

// ParseRSAPublicKey parses an RSA public key from PEM data.
func ParseRSAPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return rsaKey, nil
}

// --- JWKS ---

// JWK represents a JSON Web Key.
type JWK struct {
	KTY string `json:"kty"`
	Use string `json:"use"`
	KID string `json:"kid"`
	ALG string `json:"alg"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	CRV string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// KeyManager manages signing keys with rotation support.
type KeyManager struct {
	mu       sync.RWMutex
	keys     []*KeyPair
	activeID string
}

// NewKeyManager creates a new KeyManager with the given initial key pair.
func NewKeyManager(initial *KeyPair) *KeyManager {
	return &KeyManager{
		keys:     []*KeyPair{initial},
		activeID: initial.ID,
	}
}

// ActiveKey returns the current active signing key.
func (km *KeyManager) ActiveKey() *KeyPair {
	km.mu.RLock()
	defer km.mu.RUnlock()
	for _, kp := range km.keys {
		if kp.ID == km.activeID {
			return kp
		}
	}
	return km.keys[0]
}

// RotateKey adds a new key and makes it active while keeping old keys for verification.
func (km *KeyManager) RotateKey(newKey *KeyPair) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys = append(km.keys, newKey)
	km.activeID = newKey.ID
	// Keep only last 3 keys
	if len(km.keys) > 3 {
		km.keys = km.keys[len(km.keys)-3:]
	}
}

// GetSigningMethod returns the JWT signing method for the active key.
func (km *KeyManager) GetSigningMethod() jwt.SigningMethod {
	key := km.ActiveKey()
	switch key.Algorithm {
	case "ES256":
		return jwt.SigningMethodES256
	default:
		return jwt.SigningMethodRS256
	}
}

// GetJWKS returns the public JWKS for all managed keys.
func (km *KeyManager) GetJWKS() JWKS {
	km.mu.RLock()
	defer km.mu.RUnlock()

	jwks := JWKS{Keys: make([]JWK, 0, len(km.keys))}
	for _, kp := range km.keys {
		switch pub := kp.PublicKey.(type) {
		case *rsa.PublicKey:
			jwks.Keys = append(jwks.Keys, JWK{
				KTY: "RSA",
				Use: "sig",
				KID: kp.ID,
				ALG: kp.Algorithm,
				N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			})
		case *ecdsa.PublicKey:
			jwks.Keys = append(jwks.Keys, JWK{
				KTY: "EC",
				Use: "sig",
				KID: kp.ID,
				ALG: kp.Algorithm,
				CRV: "P-256",
				X:   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
				Y:   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
			})
		}
	}
	return jwks
}

// FindKeyByID finds a key pair by its Key ID.
func (km *KeyManager) FindKeyByID(kid string) *KeyPair {
	km.mu.RLock()
	defer km.mu.RUnlock()
	for _, kp := range km.keys {
		if kp.ID == kid {
			return kp
		}
	}
	return nil
}

// --- Token Generation ---

// GenerateRandomBytes generates cryptographically secure random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generating random bytes: %w", err)
	}
	return b, nil
}

// GenerateRandomString generates a URL-safe random string.
func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:length], nil
}

// GenerateOpaqueToken generates a 256-bit opaque token.
func GenerateOpaqueToken() (string, error) {
	b, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken produces a SHA-256 hash of a token for storage.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// TimingSafeEqual compares two strings in constant time.
func TimingSafeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HMACSign signs a message with HMAC-SHA256.
func HMACSign(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

// HMACVerify verifies an HMAC-SHA256 signature in constant time.
func HMACVerify(message, signature, secret []byte) bool {
	expected := HMACSign(message, secret)
	return hmac.Equal(expected, signature)
}
