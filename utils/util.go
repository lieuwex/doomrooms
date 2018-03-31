package utils

import (
	"crypto/rand"
	"encoding/base32"
	"os"
)

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// other error
	return false, err
}

func GenerateUID(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(bytes)[:length], nil
}
