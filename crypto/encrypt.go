package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	AESKEyLength = 32
	NonceLength  = 12
)

func (pair *KeyPair) EncryptFor(cleartext []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	// generate a random 32 bytes long key
	key := make([]byte, AESKEyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	// encrypt the key with RSA
	encKey, err := pair.EncryptBlockFor(key, pubKey)
	if err != nil {
		return nil, err
	}

	// use that key to encrypt the cleartext in AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nil, nonce, cleartext, nil)

	keySizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(keySizeBuf, uint32(len(encKey)))

	// send all
	encrypted = append(encKey, encrypted...)     // key enc
	encrypted = append(keySizeBuf, encrypted...) // ksz key enc
	encrypted = append(nonce, encrypted...)      // nonce ksz key enc

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

func (pair *KeyPair) DecryptBlock(block []byte) ([]byte, error) {
	return rsa.DecryptOAEP(
		Hasher.New(),
		rand.Reader,
		pair.Private,
		block,
		[]byte(""))
}

func (pair *KeyPair) Decrypt(ciphertext []byte) ([]byte, error) {
	dataAvailable := len(ciphertext)
	if dataAvailable < NonceLength {
		return nil, fmt.Errorf("data buffer too short")
	}

	nonce := ciphertext[0:NonceLength]
	dataAvailable -= NonceLength

	if dataAvailable < 4 {
		return nil, fmt.Errorf("data buffer too short")
	}

	keySizeBuf := ciphertext[NonceLength:NonceLength+4]
	keySize := binary.LittleEndian.Uint32(keySizeBuf)
	dataAvailable -= 4

	if dataAvailable < int(keySize) {
		return nil, fmt.Errorf("data buffer too short")
	}

	encKey := ciphertext[NonceLength + 4: NonceLength+4+keySize]
	ciphertext = ciphertext[NonceLength+4+keySize:]

	// decrypt the key
	key, err := pair.DecryptBlock(encKey)
	if err != nil {
		return nil, err
	}

	// decrypt the payload
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
