package crypto

import (
	"crypto/rand"
	"crypto/rsa"
)

// not defined for older versions of go
func pubKeySize(pub *rsa.PublicKey) int {
	return (pub.N.BitLen() + 7) / 8
}

func (pair *KeyPair) EncryptFor(cleartext []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	blockSize := pubKeySize(pubKey) - 2*Hasher.Size() - 2
	encrypted := make([]byte, 0)
	dataSize := len(cleartext)
	dataLeft := dataSize
	dataOff := 0

	for dataLeft > 0 {
		sz := blockSize
		if dataLeft < blockSize {
			sz = dataLeft
		}

		block := cleartext[dataOff : dataOff+sz]
		if encBlock, err := pair.EncryptBlockFor(block, pubKey); err != nil {
			return nil, err
		} else {
			encrypted = append(encrypted, encBlock...)
		}

		dataOff += sz
		dataLeft -= sz
	}

	return encrypted, nil
}

func (pair *KeyPair) EncryptBlockFor(block []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(
		Hasher.New(),
		rand.Reader,
		pubKey,
		block,
		[]byte(""))

}

func (pair *KeyPair) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(
		Hasher.New(),
		rand.Reader,
		pair.Private,
		ciphertext,
		[]byte(""))
}
