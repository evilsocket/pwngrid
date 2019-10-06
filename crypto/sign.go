package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

var pssOpts = rsa.PSSOptions{
	SaltLength: 16,
}

const Hasher = crypto.SHA256

func (pair *KeyPair) Sign(hash crypto.Hash, hashed []byte) ([]byte, error) {
	return rsa.SignPSS(rand.Reader, pair.Private, hash, hashed, &pssOpts)
}

func (pair *KeyPair) SignMessage(data []byte) ([]byte, error) {
	hasher := Hasher.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	return pair.Sign(Hasher, hash)
}

func (pair *KeyPair) Verify(signature []byte, hasher crypto.Hash, hash []byte) error {
	return rsa.VerifyPSS(
		pair.Public,
		hasher,
		hash,
		signature,
		&pssOpts)
}

func (pair *KeyPair) VerifyMessage(data []byte, signature []byte) error {
	hasher := Hasher.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	// log.Info("hash(data) = %x", hash)
	// log.Info("signature  = %x", signature)
	return pair.Verify(signature, Hasher, hash)
}