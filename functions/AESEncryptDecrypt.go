package functions

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	RotativeKeys sync.Map // map[string]bool
	PemanentKeys sync.Map // map[string]string
)

func GenerateAESKey() (string, error) {
	key := make([]byte, 16)
	// only utf-8 characters
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func useAESKey(key string) bool {
	_, ok := RotativeKeys.Load(key)
	if ok {
		RotativeKeys.Delete(key)
	}
	return ok
}

func EncryptAES(key string, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plaintextBytes := []byte(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintextBytes)
	return hex.EncodeToString(ciphertext), nil
}

func DecryptAES(key string, ct string) (string, error) {
	ciphertext, err := hex.DecodeString(ct)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func DecyptAESKnownKey(key string, ct string) (string, error) {
	if useAESKey(key) {
		return DecryptAES(key, ct)
	}
	return "", errors.New("Key not found")
}
