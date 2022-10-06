package utils

import (
	"crypto/sha1"
	"fmt"
	"hydra-blocking/external/config"
)

func ValidateHash(customerCode, requestHash string, config *config.Config) bool {
	h := sha1.New()
	h.Write([]byte(customerCode + config.HydraHashSumSalt))
	bs := h.Sum(nil)

	hash := fmt.Sprintf("%x", bs)
	if hash == requestHash {
		return true
	}

	return false
}
