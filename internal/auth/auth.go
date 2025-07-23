package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetAPIKey retrieves the API key from the request headers.
// It returns the API key as a string or an error if the key is not found.
// The API key is expected to be in the "Authorization" header in the format
// "Bearer <api_key>".

func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authentication header provided")
	}

	vals := strings.Split(val, " ")
	if len(vals) != 2 || vals[0] != "Bearer" {
		return "", errors.New("malformed authentication header")
	}
	return vals[1], nil
}
