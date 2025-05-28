package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var masterKey []byte

func InitCrypto() error {
	base64Key := os.Getenv("VAULT_MASTER_KEY")
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return err
	}
	if len(key) != 32 {
		return errors.New("decoded VAULT_MASTER_KEY must be 32 bytes")
	}
	masterKey = key
	return nil
}

// Encrypt slice of bytes into hashed format using cipher
func Encrypt(plain []byte) ([]byte, error) {
	if masterKey == nil {
		err := InitCrypto()
		if err != nil {
			return nil, err
		}
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aesgcm.Seal(nonce, nonce, plain, nil), nil
}

// Decrypt data from cyphered format to bytes
func Decrypt(ciphertext []byte) ([]byte, error) {
	if masterKey == nil {
		err := InitCrypto()
		if err != nil {
			return nil, err
		}
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesgcm.Open(nil, nonce, ct, nil)
}
