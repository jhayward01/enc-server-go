package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"strconv"
)

type KeyGen interface {
	GetGCMCipher(key []byte) (gcmCipher cipher.AEAD, err error)
	RandomKey() (key []byte, err error)
	RandomNonce(nonceSize int) (nonce []byte, err error)
}

type keyGenImpl struct {
	keySize int
}

func (k *keyGenImpl) GetGCMCipher(key []byte) (gcmCipher cipher.AEAD, err error) {

	// Create an AES cipher (encryption algorithm).
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Wrap cipher in GCM mode for arbitrary sized blocks and authentication.
	if gcmCipher, err = cipher.NewGCM(aesCipher); err != nil {
		return nil, err
	}

	return gcmCipher, nil
}

func (k *keyGenImpl) RandomKey() (key []byte, err error) {

	// Package crypto/rand doesn't require seeding.
	key = make([]byte, k.keySize)
	if _, err = rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (k *keyGenImpl) RandomNonce(nonceSize int) (nonce []byte, err error) {

	// Package crypto/rand doesn't require seeding.
	nonce = make([]byte, nonceSize)
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func MakeKeyGen(configs map[string]string) (k KeyGen, err error) {

	// Verify required configurations.
	if ok, missing := VerifyConfigs(configs, []string{"keySize"}); !ok {
		err = errors.New("MakeKeyGen missing configuration " + missing)
		return nil, err
	}

	// Initialize fields that require error handling.
	keySize, err := strconv.Atoi(configs["keySize"])
	if err != nil {
		return nil, err
	}

	// Build keygen implementation.
	k = &keyGenImpl{
		keySize: keySize,
	}

	return k, nil
}
