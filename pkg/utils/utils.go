package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/opencontainers/go-digest"
)

// GetContentLength returns the content length of the request.
func GetContentLength(req *http.Request) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("request is nil")
	}
	str := req.Header.Get("Content-Length")
	if str == "" {
		return 0, nil
	}
	length, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("content length is not a number")
	}
	return length, nil
}

// GenPathByDigest generates the path by the digest.
func GenPathByDigest(digest digest.Digest) string {
	hex := digest.Hex()
	return fmt.Sprintf("%s/%s/%s/%s", digest.Algorithm(), hex[0:2], hex[2:4], hex[4:])
}
