package helper

// This file is currently unused, traffic encryption will be implemented in later updates.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// Encryption key to use for cryptography and send to the server.
var EncryptionKey = GenerateEncryptionKey()

func GenerateEncryptionKey() []byte {
	key := make([]byte, 24)
	rand.Read(key)

	return key
}

func Encrypt(plaindata []byte) ([]byte, error) {
	block, err := aes.NewCipher(EncryptionKey)
	if err != nil {
		return nil, err
	}

	// Create a new GCM block cipher with the given key.
	// GCM provides authenticated encryption, which ensures both confidentiality and integrity.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a new nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the original data using the GCM block cipher.
	ciphertext := gcm.Seal(nonce, nonce, plaindata, nil)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(EncryptionKey)
	if err != nil {
		return nil, err
	}

	// Create a new GCM block cipher with the given key
	// GCM provides authenticated encryption, which ensures both confidentiality and integrity
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()

	// Extract the nonce from the ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext using the GCM block cipher
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
