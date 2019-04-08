package kops

import (
	"fmt"
	"github.com/infinityworks/fk-infra/executable"
	"github.com/infinityworks/fk-infra/util"
	"log"
	"os"
	"runtime"
)

const (
	kopsBinaryLocation = ".fk-infra/kops"
)

func ExecuteKops(args ...string) []byte {
	return executable.CacheOrDownload(kopsBinaryLocation,
		func() string {
			downloadUrl := fmt.Sprintf("https://github.com/kubernetes/kops/releases/download/1.11.1/kops-%s-%s", runtime.GOOS, runtime.GOARCH)
			log.Printf("Downloading kops from %s", downloadUrl)
			return downloadUrl
		},
		func(tempBinaryLocation string) {
			util.CheckError(os.Rename(tempBinaryLocation, kopsBinaryLocation))
			util.CheckError(os.Chmod(kopsBinaryLocation, 0740))
		},
		args...)
}
