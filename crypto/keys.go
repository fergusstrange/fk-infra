package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/infinityworks/fk-infra/util"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
)

const (
	PublicKeyFile = "keys/public_key.pub"

	keyDir         = "keys"
	privateKeyFile = "keys/private_key"
	gitignoreFile  = "keys/.gitignore"
)

func CreateOrValidateExistingKey() {
	if !util.PathExists(keyDir) {
		log.Println("fresh installation, creating keys directory")
		createKeyDirectory()
	}

	if util.PathExists(encryptedName(privateKeyFile)) {
		log.Println("existing key detected and being used")
		DecryptKeys()
	} else {
		log.Println("creating new key")
		privateKey := createAndEncryptPrivateKey()
		createAndEncryptPublicKey(privateKey)
	}

	privateKeyBytes, err := ioutil.ReadFile(privateKeyFile)
	util.CheckError(err)

	_, err = ssh.ParseRawPrivateKey(privateKeyBytes)
	util.CheckError(err)
}

func DecryptKeys() {
	Decrypt(encryptedName(privateKeyFile))
	Decrypt(encryptedName(PublicKeyFile))
}

func encryptedName(fileName string) string {
	return fmt.Sprintf("%s.enc", fileName)
}

func createAndEncryptPrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	util.CheckError(err)

	buffer := new(bytes.Buffer)

	util.CheckError(pem.Encode(buffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}))

	util.WriteFile(privateKeyFile, buffer.Bytes())

	Encrypt(privateKeyFile)

	return privateKey
}

func createAndEncryptPublicKey(privateKey *rsa.PrivateKey) {
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	util.CheckError(err)
	util.WriteFile(PublicKeyFile, ssh.MarshalAuthorizedKey(publicKey))

	Encrypt(PublicKeyFile)
}

func createKeyDirectory() {
	util.CheckError(os.MkdirAll(keyDir, os.FileMode(0777)))
	createGitIgnore()
}

func createGitIgnore() {
	util.WriteFile(gitignoreFile, []byte(
		`*
!*.enc
!*.gitignore
`))
}
