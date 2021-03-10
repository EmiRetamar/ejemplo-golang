package crypto

import (
	"errors"
)

type EcoCryptoType int

const (
	AES EcoCryptoType = 1
)

type EcoCrypto interface {
	Encode(any string) (string, error)
	Decode(any string) (string, error)
}

func NewCrypto(typeCrypto EcoCryptoType) (EcoCrypto, error) {

	switch typeCrypto {
		case AES:
			return newAesCrypto()
		default:
			return nil, errors.New("no implment crypto type")
	}
}
