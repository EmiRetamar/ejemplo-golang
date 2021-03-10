package crypto

import (
	"encoding/hex"
	"errors"
	"crypto/aes"
	"io"
	"crypto/rand"
	"crypto/cipher"
	"fmt"
)

const aesKey = "6573746f2e65732e756e612e6b65792e706172612e616573"

var instance *aesCrypto

type aesCrypto struct {
	nonce []byte
}

func (a *aesCrypto) Encode(any string) (string, error){

	if len(any) == 0 {
		return "", errors.New("encode is missing")
	}

	key, _ := hex.DecodeString(aesKey)
	textToEncode := []byte(any)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(textToEncode))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], textToEncode)

	return fmt.Sprintf("%x", ciphertext), nil
}

func (a *aesCrypto) Decode(any string) (string, error){

	if len(any) == 0 {
		return "", errors.New("decode is missing")
	}

	key, _ := hex.DecodeString(aesKey)
	ciphertext, _ := hex.DecodeString(any)

	block, err := aes.NewCipher(key)
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

func (a *aesCrypto) init() error {

	a.nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, a.nonce); err != nil {
		return err
	}
	return nil
}

func newAesCrypto() (EcoCrypto, error) {
	if instance != nil {
		return instance, nil
	}
	instance := new(aesCrypto)
	if err := instance.init(); err != nil {
		return nil, err
	}
	return instance, nil
}
