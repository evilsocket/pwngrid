package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type KeyPair struct {
	Path        string
	Bits        int
	PrivatePath string
	Private     *rsa.PrivateKey
	PrivatePEM  []byte
	PublicPath  string
	Public      *rsa.PublicKey
	PublicPEM   []byte
	// sha256 of PublicSSH
	Fingerprint    []byte
	FingerprintHex string
}

func pubKeyToPEM(key *rsa.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: bytes,
		},
	), nil
}

func FromPublicPEM(pubPEM string) (pair *KeyPair, err error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pair = &KeyPair{}
	ok := false

	if pair.Public, ok = pub.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("not an RSA key")
	}

	return pair, pair.setupPublic()
}

func PrivatePath(keysPath string) string {
	return path.Join(keysPath, "id_rsa")
}

func Load(keysPath string) (pair *KeyPair, err error) {
	privFile := PrivatePath(keysPath)
	pair = &KeyPair{
		Path:        keysPath,
		PrivatePath: privFile,
		PublicPath:  privFile + ".pub",
	}
	return pair, pair.Load()
}

func KeysExist(keysPath string) bool {
	return fs.Exists(keysPath) && fs.Exists(PrivatePath(keysPath))
}

func LoadOrCreate(keysPath string, bits int) (pair *KeyPair, err error) {
	privFile := PrivatePath(keysPath)
	pair = &KeyPair{
		Path:        keysPath,
		Bits:        bits,
		PrivatePath: privFile,
		PublicPath:  privFile + ".pub",
	}

	if !fs.Exists(pair.PrivatePath) {
		if !fs.Exists(keysPath) {
			log.Debug("creating %s", keysPath)
			if err := os.MkdirAll(keysPath, os.ModePerm); err != nil {
				return nil, fmt.Errorf("could not create %s: %v", keysPath, err)
			}
		}
		log.Info("%s not found, generating keypair ...", pair.PrivatePath)

		if pair.Private, err = rsa.GenerateKey(rand.Reader, bits); err != nil {
			return nil, fmt.Errorf("could not generate private key: %v", err)
		}
		pair.Public = &pair.Private.PublicKey

		if err = pair.Save(); err != nil {
			return nil, fmt.Errorf("could not save keypair: %v", err)
		}
	} else if err = pair.Load(); err != nil {
		return nil, fmt.Errorf("could not load keypair: %v", err)
	}

	return pair, nil
}

func (pair *KeyPair) setupPublic() (err error) {
	if pair.PublicPEM, err = pubKeyToPEM(pair.Public); err != nil {
		return fmt.Errorf("failed converting public key to PEM: %v", err)
	}

	cleanPEM := strings.TrimRight(string(pair.PublicPEM), "\n")

	hash := Hasher.New()
	hash.Write([]byte(cleanPEM))

	pair.Fingerprint = hash.Sum(nil)
	pair.FingerprintHex = fmt.Sprintf("%02x", pair.Fingerprint)

	return nil
}

func (pair *KeyPair) Save() (err error) {
	prvKeyBytes := x509.MarshalPKCS1PrivateKey(pair.Private)
	pair.PrivatePEM = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: prvKeyBytes,
		},
	)

	if err = ioutil.WriteFile(pair.PrivatePath, pair.PrivatePEM, os.ModePerm); err != nil {
		return
	}

	log.Debug("%s created", pair.PrivatePath)

	if err = pair.setupPublic(); err != nil {
		return err
	}

	err = ioutil.WriteFile(pair.PublicPath, pair.PublicPEM, os.ModePerm)

	log.Debug("%s created", pair.PublicPath)
	return
}

func (pair *KeyPair) Load() (err error) {
	log.Debug("reading %s ...", pair.PrivatePath)
	if pair.PrivatePEM, err = ioutil.ReadFile(pair.PrivatePath); err != nil {
		return
	}

	block, _ := pem.Decode(pair.PrivatePEM)
	if block == nil {
		return fmt.Errorf("failed decoding PEM from %s", pair.PrivatePath)
	}

	if pair.Private, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		return fmt.Errorf("failed parsing %s: %v", pair.PrivatePath, err)
	}

	pair.Public = &pair.Private.PublicKey
	return pair.setupPublic()
}
