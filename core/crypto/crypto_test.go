package crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/hex"
	"strings"
	"testing"
)

// --- Argon2id Password Hashing ---

func TestHashPassword(t *testing.T) {
	params := Argon2Params{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}

	hash, err := HashPassword("correcthorse", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("hash does not start with expected prefix: %s", hash)
	}

	// Ensure the hash has all expected parts
	parts := strings.Split(hash, "$")
	// Expected: empty, argon2id, v=19, m=..., salt, hash -> 6 parts
	if len(parts) < 6 {
		t.Errorf("expected at least 6 parts in hash, got %d: %s", len(parts), hash)
	}
}

func TestHashPassword_DifferentPasswords(t *testing.T) {
	params := DefaultArgon2Params()

	hash1, err := HashPassword("password1", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	hash2, err := HashPassword("password2", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("different passwords produced the same hash")
	}
}

func TestHashPassword_SamePasswordDifferentSalts(t *testing.T) {
	params := DefaultArgon2Params()

	hash1, err := HashPassword("samepassword", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	hash2, err := HashPassword("samepassword", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("same password should produce different hashes due to random salt")
	}
}

func TestVerifyPassword(t *testing.T) {
	params := DefaultArgon2Params()

	password := "MyS3cureP@ss!"
	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !match {
		t.Error("VerifyPassword should return true for the correct password")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	params := DefaultArgon2Params()

	hash, err := HashPassword("correct", params)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := VerifyPassword("wrong", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if match {
		t.Error("VerifyPassword should return false for the wrong password")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	_, err := VerifyPassword("password", "not-a-valid-hash")
	if err == nil {
		t.Error("VerifyPassword should return error for invalid hash format")
	}
}

// --- AES-256-GCM Encryption ---

func testKey() string {
	// 32 bytes = 64 hex chars
	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

func TestEncryptDecrypt(t *testing.T) {
	key := testKey()
	plaintext := []byte("Hello, World!")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	if string(ciphertext) == string(plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt returned %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestEncrypt_DifferentKeys(t *testing.T) {
	key1 := testKey()
	key2 := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	plaintext := []byte("secret data")

	ct1, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt with key1 returned error: %v", err)
	}

	ct2, err := Encrypt(plaintext, key2)
	if err != nil {
		t.Fatalf("Encrypt with key2 returned error: %v", err)
	}

	// Different keys + different nonces should produce different ciphertexts
	if string(ct1) == string(ct2) {
		t.Error("ciphertexts with different keys should differ")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := testKey()
	key2 := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	ciphertext, err := Encrypt([]byte("hello"), key1)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("Decrypt with wrong key should return an error")
	}
}

func TestEncrypt_InvalidKey(t *testing.T) {
	// Key too short
	_, err := Encrypt([]byte("data"), "0123456789abcdef")
	if err == nil {
		t.Error("Encrypt should return error for key that is not 32 bytes")
	}

	// Key is not valid hex
	_, err = Encrypt([]byte("data"), "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	if err == nil {
		t.Error("Encrypt should return error for non-hex key")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	key := testKey()
	_, err := Decrypt([]byte("short"), key)
	if err == nil {
		t.Error("Decrypt should return error for too-short ciphertext")
	}
}

// --- RSA/ECDSA Key Generation ---

func TestGenerateRSAKeyPair(t *testing.T) {
	kp, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair returned error: %v", err)
	}

	if kp.ID == "" {
		t.Error("KeyPair ID should not be empty")
	}
	if kp.Algorithm != "RS256" {
		t.Errorf("Algorithm = %q, want RS256", kp.Algorithm)
	}

	if _, ok := kp.PrivateKey.(*rsa.PrivateKey); !ok {
		t.Error("PrivateKey should be *rsa.PrivateKey")
	}
	if _, ok := kp.PublicKey.(*rsa.PublicKey); !ok {
		t.Error("PublicKey should be *rsa.PublicKey")
	}
}

func TestGenerateRSAKeyPair_MinBits(t *testing.T) {
	// Requesting fewer than 2048 bits should be upgraded to 2048
	kp, err := GenerateRSAKeyPair(1024)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair returned error: %v", err)
	}

	privKey := kp.PrivateKey.(*rsa.PrivateKey)
	if privKey.N.BitLen() < 2048 {
		t.Errorf("RSA key should be at least 2048 bits, got %d", privKey.N.BitLen())
	}
}

func TestGenerateECDSAKeyPair(t *testing.T) {
	kp, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateECDSAKeyPair returned error: %v", err)
	}

	if kp.ID == "" {
		t.Error("KeyPair ID should not be empty")
	}
	if kp.Algorithm != "ES256" {
		t.Errorf("Algorithm = %q, want ES256", kp.Algorithm)
	}

	if _, ok := kp.PrivateKey.(*ecdsa.PrivateKey); !ok {
		t.Error("PrivateKey should be *ecdsa.PrivateKey")
	}
	if _, ok := kp.PublicKey.(*ecdsa.PublicKey); !ok {
		t.Error("PublicKey should be *ecdsa.PublicKey")
	}
}

// --- Random Token Generation ---

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 8", 8},
		{"length 16", 16},
		{"length 32", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomString returned error: %v", err)
			}
			if len(s) != tt.length {
				t.Errorf("got string of length %d, want %d", len(s), tt.length)
			}
		})
	}
}

func TestGenerateRandomString_Unique(t *testing.T) {
	s1, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString returned error: %v", err)
	}
	s2, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString returned error: %v", err)
	}
	if s1 == s2 {
		t.Error("two random strings should be different")
	}
}

func TestGenerateOpaqueToken(t *testing.T) {
	token, err := GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("GenerateOpaqueToken returned error: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	// 32 bytes in base64 raw URL encoding = 43 chars
	if len(token) < 40 {
		t.Errorf("token seems too short: %d chars", len(token))
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	b, err := GenerateRandomBytes(32)
	if err != nil {
		t.Fatalf("GenerateRandomBytes returned error: %v", err)
	}
	if len(b) != 32 {
		t.Errorf("got %d bytes, want 32", len(b))
	}
}

// --- Token Hashing ---

func TestHashToken(t *testing.T) {
	token := "my-secret-token"
	hash := HashToken(token)

	if hash == "" {
		t.Error("hash should not be empty")
	}
	if hash == token {
		t.Error("hash should differ from the original token")
	}

	// Should be a valid hex string (SHA-256 = 64 hex chars)
	if len(hash) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(hash))
	}
	_, err := hex.DecodeString(hash)
	if err != nil {
		t.Errorf("hash is not valid hex: %v", err)
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "deterministic-token"
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Error("HashToken should be deterministic")
	}
}

func TestHashToken_DifferentInputs(t *testing.T) {
	h1 := HashToken("token-a")
	h2 := HashToken("token-b")
	if h1 == h2 {
		t.Error("different tokens should produce different hashes")
	}
}

// --- HMAC ---

func TestHMACSign(t *testing.T) {
	message := []byte("hello world")
	secret := []byte("my-secret-key")

	sig := HMACSign(message, secret)
	if len(sig) == 0 {
		t.Error("signature should not be empty")
	}
	// SHA-256 HMAC produces 32 bytes
	if len(sig) != 32 {
		t.Errorf("expected 32 byte signature, got %d", len(sig))
	}
}

func TestHMACSign_Deterministic(t *testing.T) {
	message := []byte("same message")
	secret := []byte("same secret")

	sig1 := HMACSign(message, secret)
	sig2 := HMACSign(message, secret)

	if string(sig1) != string(sig2) {
		t.Error("HMACSign should be deterministic for same inputs")
	}
}

func TestHMACVerify(t *testing.T) {
	message := []byte("verify me")
	secret := []byte("key123")

	sig := HMACSign(message, secret)
	if !HMACVerify(message, sig, secret) {
		t.Error("HMACVerify should return true for valid signature")
	}
}

func TestHMACVerify_WrongSignature(t *testing.T) {
	message := []byte("hello")
	secret := []byte("key")

	wrongSig := []byte("this-is-not-a-valid-signature-xx")
	if HMACVerify(message, wrongSig, secret) {
		t.Error("HMACVerify should return false for wrong signature")
	}
}

func TestHMACVerify_WrongMessage(t *testing.T) {
	secret := []byte("key")
	sig := HMACSign([]byte("original"), secret)

	if HMACVerify([]byte("tampered"), sig, secret) {
		t.Error("HMACVerify should return false when message is tampered")
	}
}

func TestHMACVerify_WrongSecret(t *testing.T) {
	message := []byte("data")
	sig := HMACSign(message, []byte("correct-secret"))

	if HMACVerify(message, sig, []byte("wrong-secret")) {
		t.Error("HMACVerify should return false with wrong secret")
	}
}

// --- Timing Safe Equal ---

func TestTimingSafeEqual(t *testing.T) {
	if !TimingSafeEqual("abc", "abc") {
		t.Error("TimingSafeEqual should return true for equal strings")
	}
	if TimingSafeEqual("abc", "xyz") {
		t.Error("TimingSafeEqual should return false for different strings")
	}
	if TimingSafeEqual("abc", "ab") {
		t.Error("TimingSafeEqual should return false for different length strings")
	}
	if TimingSafeEqual("", "a") {
		t.Error("TimingSafeEqual should return false when one is empty")
	}
	if !TimingSafeEqual("", "") {
		t.Error("TimingSafeEqual should return true for two empty strings")
	}
}

// --- KeyManager ---

func TestKeyManager_ActiveKey(t *testing.T) {
	kp, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair returned error: %v", err)
	}

	km := NewKeyManager(kp)
	active := km.ActiveKey()

	if active.ID != kp.ID {
		t.Errorf("ActiveKey ID = %q, want %q", active.ID, kp.ID)
	}
	if active.Algorithm != kp.Algorithm {
		t.Errorf("ActiveKey Algorithm = %q, want %q", active.Algorithm, kp.Algorithm)
	}
}

func TestKeyManager_RotateKey(t *testing.T) {
	kp1, _ := GenerateRSAKeyPair(2048)
	kp2, _ := GenerateRSAKeyPair(2048)

	km := NewKeyManager(kp1)
	km.RotateKey(kp2)

	active := km.ActiveKey()
	if active.ID != kp2.ID {
		t.Errorf("after rotation, ActiveKey ID = %q, want %q", active.ID, kp2.ID)
	}

	// Old key should still be findable
	found := km.FindKeyByID(kp1.ID)
	if found == nil {
		t.Error("old key should still be findable after rotation")
	}
}

func TestKeyManager_RotateKey_MaxThreeKeys(t *testing.T) {
	kp1, _ := GenerateRSAKeyPair(2048)
	kp2, _ := GenerateRSAKeyPair(2048)
	kp3, _ := GenerateRSAKeyPair(2048)
	kp4, _ := GenerateRSAKeyPair(2048)

	km := NewKeyManager(kp1)
	km.RotateKey(kp2)
	km.RotateKey(kp3)
	km.RotateKey(kp4)

	// kp1 should have been evicted (only last 3 kept)
	if km.FindKeyByID(kp1.ID) != nil {
		t.Error("oldest key should have been evicted after 4 rotations")
	}
	// kp2, kp3, kp4 should still exist
	if km.FindKeyByID(kp2.ID) == nil {
		t.Error("kp2 should still exist")
	}
	if km.FindKeyByID(kp3.ID) == nil {
		t.Error("kp3 should still exist")
	}
	if km.FindKeyByID(kp4.ID) == nil {
		t.Error("kp4 should still exist")
	}
}

func TestKeyManager_FindKeyByID_NotFound(t *testing.T) {
	kp, _ := GenerateRSAKeyPair(2048)
	km := NewKeyManager(kp)

	if km.FindKeyByID("nonexistent") != nil {
		t.Error("FindKeyByID should return nil for unknown key ID")
	}
}

func TestKeyManager_GetJWKS_RSA(t *testing.T) {
	kp, _ := GenerateRSAKeyPair(2048)
	km := NewKeyManager(kp)

	jwks := km.GetJWKS()
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key in JWKS, got %d", len(jwks.Keys))
	}

	key := jwks.Keys[0]
	if key.KTY != "RSA" {
		t.Errorf("KTY = %q, want RSA", key.KTY)
	}
	if key.Use != "sig" {
		t.Errorf("Use = %q, want sig", key.Use)
	}
	if key.KID != kp.ID {
		t.Errorf("KID = %q, want %q", key.KID, kp.ID)
	}
	if key.ALG != "RS256" {
		t.Errorf("ALG = %q, want RS256", key.ALG)
	}
	if key.N == "" {
		t.Error("N should not be empty for RSA key")
	}
	if key.E == "" {
		t.Error("E should not be empty for RSA key")
	}
}

func TestKeyManager_GetJWKS_ECDSA(t *testing.T) {
	kp, _ := GenerateECDSAKeyPair()
	km := NewKeyManager(kp)

	jwks := km.GetJWKS()
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key in JWKS, got %d", len(jwks.Keys))
	}

	key := jwks.Keys[0]
	if key.KTY != "EC" {
		t.Errorf("KTY = %q, want EC", key.KTY)
	}
	if key.ALG != "ES256" {
		t.Errorf("ALG = %q, want ES256", key.ALG)
	}
	if key.CRV != "P-256" {
		t.Errorf("CRV = %q, want P-256", key.CRV)
	}
	if key.X == "" {
		t.Error("X should not be empty for EC key")
	}
	if key.Y == "" {
		t.Error("Y should not be empty for EC key")
	}
}

func TestKeyManager_GetJWKS_MultipleKeys(t *testing.T) {
	kp1, _ := GenerateRSAKeyPair(2048)
	kp2, _ := GenerateECDSAKeyPair()

	km := NewKeyManager(kp1)
	km.RotateKey(kp2)

	jwks := km.GetJWKS()
	if len(jwks.Keys) != 2 {
		t.Fatalf("expected 2 keys in JWKS, got %d", len(jwks.Keys))
	}
}

func TestKeyManager_GetSigningMethod_RSA(t *testing.T) {
	kp, _ := GenerateRSAKeyPair(2048)
	km := NewKeyManager(kp)

	sm := km.GetSigningMethod()
	if sm.Alg() != "RS256" {
		t.Errorf("signing method = %q, want RS256", sm.Alg())
	}
}

func TestKeyManager_GetSigningMethod_ECDSA(t *testing.T) {
	kp, _ := GenerateECDSAKeyPair()
	km := NewKeyManager(kp)

	sm := km.GetSigningMethod()
	if sm.Alg() != "ES256" {
		t.Errorf("signing method = %q, want ES256", sm.Alg())
	}
}
