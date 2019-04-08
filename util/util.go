package util

import (
	"github.com/spf13/afero"
	"io/ioutil"
	"log"
	"math/rand"
)

const alphaNumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func RandomAlphaNumeric(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphaNumericChars[rand.Intn(len(alphaNumericChars))]
	}
	return string(b)
}

func WriteFile(fileName string, content []byte) {
	CheckError(ioutil.WriteFile(fileName, content, 0666))
}

func String(str string) *string {
	return &str
}

func Int32(num int32) *int32 {
	return &num
}

func PathExists(path string) bool {
	directoryExists, err := afero.Exists(afero.NewOsFs(), path)
	return err == nil && directoryExists
}