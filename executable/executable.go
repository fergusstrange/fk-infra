package executable

import (
	"bytes"
	"github.com/infinityworks/fk-infra/util"
	"github.com/spf13/afero"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	tempBinaryLocation = ".fk-infra/temp_binary"
)

func CacheOrDownload(
	binaryLocation string,
	downloadLocation func() string,
	postDownloadFunction func(tempBinaryLocation string),
	args ...string) []byte {
	if exists, err := afero.Exists(afero.NewOsFs(), binaryLocation); err == nil && exists {
		return execute(binaryLocation, args...)
	} else {
		util.CheckError(os.MkdirAll(".fk-infra", 0750))

		resp, err := http.Get(downloadLocation())
		util.CheckError(err)

		out, err := os.Create(tempBinaryLocation)
		util.CheckError(err)

		defer func() {
			util.CheckError(resp.Body.Close())
			util.CheckError(out.Close())
		}()

		_, err = io.Copy(out, resp.Body)
		util.CheckError(err)

		postDownloadFunction(tempBinaryLocation)

		return execute(binaryLocation, args...)
	}
}

func execute(binaryLocation string, args ...string) []byte {
	log.Printf("Executing %s %s", binaryLocation, args)
	cmd := exec.Command(binaryLocation, args...)
	cmd.Stderr = os.Stderr
	dualWriter := DualWriter{
		buffer: new(bytes.Buffer),
	}
	cmd.Stdout = dualWriter
	err := cmd.Run()
	util.CheckError(err)
	return dualWriter.Bytes()
}

type DualWriter struct {
	buffer *bytes.Buffer
}

func (dualWriter DualWriter) Write(p []byte) (n int, err error) {
	_, err = os.Stdout.Write(p)
	util.CheckError(err)
	return dualWriter.buffer.Write(p)
}

func (dualWriter DualWriter) Bytes() []byte {
	return dualWriter.buffer.Bytes()
}
