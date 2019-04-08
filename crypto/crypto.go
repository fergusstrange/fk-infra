package crypto

import (
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/util"
	"github.com/steinfletcher/kms-secrets/compress"
	"github.com/steinfletcher/kms-secrets/kms"
	"github.com/steinfletcher/kms-secrets/secrets"
	"log"
	"os"
)

func Encrypt(filename string) {
	log.Printf("encrypting file %s", filename)
	config := model.FetchConfig().Spec
	newKms := kms.NewKms(config.EncryptionKey, config.Region, os.Getenv("AWS_PROFILE"))
	util.CheckError(secrets.NewSecrets(newKms, compress.NewGzipCompressor(), filename).Encrypt("./"))
}

func Decrypt(filename string) {
	log.Printf("decrypting file %s", filename)
	config := model.FetchConfig().Spec
	newKms := kms.NewKms(config.EncryptionKey, config.Region, os.Getenv("AWS_PROFILE"))
	util.CheckError(secrets.NewSecrets(newKms, compress.NewGzipCompressor(), filename).Decrypt("./"))
}
