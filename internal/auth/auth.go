package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authorization header")
	}

	// expected api key: ApiKey {value}
	vals := strings.Split(val, " ")
	if len(vals) != 2 {
		return "", errors.New("invalid api key format")
	}
	if vals[0] != "ApiKey" {
		return "", errors.New("invalid api key format")
	}

	return vals[1], nil
}
