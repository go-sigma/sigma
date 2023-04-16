package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

// Compress compresses the given string using gzip.
func Compress(src string) ([]byte, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			log.Warn().Err(err).Msg("Close file failed")
		}
	}()

	var dst bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&dst, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(gzipWriter, srcFile)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return dst.Bytes(), nil
}
