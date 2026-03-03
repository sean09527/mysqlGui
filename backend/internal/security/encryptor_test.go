package security

import (
	"strings"
	"testing"
)

// TestNewEncryptor tests creating a new Encryptor instance
func TestNewEncryptor(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	if encryptor == nil {
		t.Fatal("Expected encryptor to be created")
	}

	if len(encryptor.key) != 32 {
		t.Errorf("Expected key length 32 bytes (AES-256), got %d", len(encryptor.key))
	}
}

// TestEncryptDecrypt tests basic encryption and decryption
func TestEncryptDecrypt(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	plaintext := "my secret password"

	// Encrypt
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if ciphertext == "" {
		t.Fatal("Expected non-empty ciphertext")
	}

	if ciphertext == plaintext {
		t.Error("Ciphertext should not equal plaintext")
	}

	// Decrypt
	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", plaintext, decrypted)
	}
}

// TestEncryptEmptyString tests encrypting an empty string
func TestEncryptEmptyString(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	ciphertext, err := encryptor.Encrypt("")
	if err != nil {
		t.Fatalf("Failed to encrypt empty string: %v", err)
	}

	if ciphertext != "" {
		t.Error("Expected empty ciphertext for empty plaintext")
	}

	// Decrypt empty string
	decrypted, err := encryptor.Decrypt("")
	if err != nil {
		t.Fatalf("Failed to decrypt empty string: %v", err)
	}

	if decrypted != "" {
		t.Error("Expected empty plaintext for empty ciphertext")
	}
}

// TestEncryptDifferentPassphrases tests that different passphrases produce different keys
func TestEncryptDifferentPassphrases(t *testing.T) {
	encryptor1 := NewEncryptor("passphrase1")
	encryptor2 := NewEncryptor("passphrase2")

	plaintext := "secret data"

	// Encrypt with first encryptor
	ciphertext1, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt with encryptor1: %v", err)
	}

	// Try to decrypt with second encryptor (should fail)
	_, err = encryptor2.Decrypt(ciphertext1)
	if err == nil {
		t.Error("Expected error when decrypting with wrong passphrase")
	}
}

// TestEncryptDeterministic tests that encryption is non-deterministic (uses random nonce)
func TestEncryptNonDeterministic(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	plaintext := "same plaintext"

	// Encrypt twice
	ciphertext1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt first time: %v", err)
	}

	ciphertext2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt second time: %v", err)
	}

	// Ciphertexts should be different due to random nonce
	if ciphertext1 == ciphertext2 {
		t.Error("Expected different ciphertexts for same plaintext (non-deterministic encryption)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := encryptor.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("Failed to decrypt first ciphertext: %v", err)
	}

	decrypted2, err := encryptor.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Failed to decrypt second ciphertext: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to the same plaintext")
	}
}

// TestEncryptLongString tests encrypting a long string
func TestEncryptLongString(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	// Create a long string (1000 characters)
	plaintext := strings.Repeat("This is a long string for testing encryption. ", 20)

	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt long string: %v", err)
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt long string: %v", err)
	}

	if decrypted != plaintext {
		t.Error("Decrypted long string does not match original")
	}
}

// TestEncryptSpecialCharacters tests encrypting strings with special characters
func TestEncryptSpecialCharacters(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	testCases := []string{
		"password with spaces",
		"password!@#$%^&*()",
		"密码中文字符",
		"пароль кириллица",
		"🔒🔑 emoji password",
		"line1\nline2\ttab",
	}

	for _, plaintext := range testCases {
		ciphertext, err := encryptor.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Failed to encrypt '%s': %v", plaintext, err)
		}

		decrypted, err := encryptor.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Failed to decrypt '%s': %v", plaintext, err)
		}

		if decrypted != plaintext {
			t.Errorf("Expected '%s', got '%s'", plaintext, decrypted)
		}
	}
}

// TestDecryptInvalidCiphertext tests decrypting invalid ciphertext
func TestDecryptInvalidCiphertext(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	testCases := []string{
		"not-base64-encoded!@#",
		"dGhpcyBpcyBub3QgdmFsaWQgY2lwaGVydGV4dA==", // valid base64 but invalid ciphertext
		"YQ==",                                       // too short
	}

	for _, ciphertext := range testCases {
		_, err := encryptor.Decrypt(ciphertext)
		if err == nil {
			t.Errorf("Expected error when decrypting invalid ciphertext '%s'", ciphertext)
		}
	}
}

// TestEncryptBytesDecryptBytes tests byte encryption and decryption
func TestEncryptBytesDecryptBytes(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	plaintext := []byte("binary data: \x00\x01\x02\x03\xff")

	// Encrypt
	ciphertext, err := encryptor.EncryptBytes(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt bytes: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("Expected non-empty ciphertext")
	}

	// Decrypt
	decrypted, err := encryptor.DecryptBytes(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt bytes: %v", err)
	}

	if len(decrypted) != len(plaintext) {
		t.Errorf("Expected decrypted length %d, got %d", len(plaintext), len(decrypted))
	}

	for i := range plaintext {
		if decrypted[i] != plaintext[i] {
			t.Errorf("Byte mismatch at index %d: expected %x, got %x", i, plaintext[i], decrypted[i])
		}
	}
}

// TestEncryptBytesEmpty tests encrypting empty byte slice
func TestEncryptBytesEmpty(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	ciphertext, err := encryptor.EncryptBytes([]byte{})
	if err != nil {
		t.Fatalf("Failed to encrypt empty bytes: %v", err)
	}

	if ciphertext != nil {
		t.Error("Expected nil ciphertext for empty plaintext")
	}

	// Decrypt empty bytes
	decrypted, err := encryptor.DecryptBytes(nil)
	if err != nil {
		t.Fatalf("Failed to decrypt empty bytes: %v", err)
	}

	if decrypted != nil {
		t.Error("Expected nil plaintext for empty ciphertext")
	}
}

// TestEncryptBytesLarge tests encrypting large byte data
func TestEncryptBytesLarge(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	// Create 10KB of data
	plaintext := make([]byte, 10*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := encryptor.EncryptBytes(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt large bytes: %v", err)
	}

	decrypted, err := encryptor.DecryptBytes(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt large bytes: %v", err)
	}

	if len(decrypted) != len(plaintext) {
		t.Errorf("Expected decrypted length %d, got %d", len(plaintext), len(decrypted))
	}

	for i := range plaintext {
		if decrypted[i] != plaintext[i] {
			t.Errorf("Byte mismatch at index %d", i)
			break
		}
	}
}

// TestDecryptBytesInvalid tests decrypting invalid byte ciphertext
func TestDecryptBytesInvalid(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	testCases := [][]byte{
		[]byte("too short"),
		[]byte{0x00, 0x01, 0x02}, // too short
	}

	for _, ciphertext := range testCases {
		_, err := encryptor.DecryptBytes(ciphertext)
		if err == nil {
			t.Error("Expected error when decrypting invalid byte ciphertext")
		}
	}
}

// TestKeyDerivation tests that the same passphrase produces the same key
func TestKeyDerivation(t *testing.T) {
	encryptor1 := NewEncryptor("same-passphrase")
	encryptor2 := NewEncryptor("same-passphrase")

	// Keys should be identical
	if len(encryptor1.key) != len(encryptor2.key) {
		t.Error("Keys should have the same length")
	}

	for i := range encryptor1.key {
		if encryptor1.key[i] != encryptor2.key[i] {
			t.Error("Keys should be identical for the same passphrase")
			break
		}
	}

	// Verify they can decrypt each other's ciphertexts
	plaintext := "test data"

	ciphertext1, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt with encryptor1: %v", err)
	}

	decrypted, err := encryptor2.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("Failed to decrypt with encryptor2: %v", err)
	}

	if decrypted != plaintext {
		t.Error("Encryptor2 should be able to decrypt encryptor1's ciphertext")
	}
}

// TestEncryptionIntegrity tests that tampering with ciphertext is detected
func TestEncryptionIntegrity(t *testing.T) {
	encryptor := NewEncryptor("test-passphrase")

	plaintext := "important data"

	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Tamper with the ciphertext (flip a bit in the middle)
	tamperedBytes := []byte(ciphertext)
	if len(tamperedBytes) > 10 {
		tamperedBytes[10] ^= 0x01 // flip one bit
	}
	tamperedCiphertext := string(tamperedBytes)

	// Decryption should fail due to authentication tag mismatch
	_, err = encryptor.Decrypt(tamperedCiphertext)
	if err == nil {
		t.Error("Expected error when decrypting tampered ciphertext (GCM should detect tampering)")
	}
}
