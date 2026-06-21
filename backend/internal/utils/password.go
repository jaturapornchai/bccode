package utils

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Util struct{}

// bcryptCost resolves the bcrypt cost factor. Default is 10 (fast on modern CPUs,
// ~50ms). The legacy value was 14 (~700ms) which made every login painfully slow.
// Override with env BC_BCRYPT_COST for stricter deployments.
func bcryptCost() int {
	raw := strings.TrimSpace(os.Getenv("BC_BCRYPT_COST"))
	if raw != "" {
		if cost, err := strconv.Atoi(raw); err == nil && cost >= bcrypt.MinCost && cost <= bcrypt.MaxCost {
			return cost
		}
	}
	return 10
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost())
	return string(bytes), err
}

func CheckHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
