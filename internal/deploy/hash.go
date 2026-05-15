package deploy

import (
	"crypto/sha256"
	"fmt"
)

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum)
}
