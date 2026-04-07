package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func GetSignature(timestamp time.Time, shopID uuid.UUID, key string) string {
	hashSum := sha256.Sum256([]byte(
		fmt.Sprintf("%d:%s:%s", timestamp.UnixMilli(), shopID.String(), key),
	))

	return hex.EncodeToString(hashSum[:])
}

func GetActionSignature(timestamp time.Time, transactionID uuid.UUID, key string) string {
	hashSum := sha256.Sum256([]byte(
		fmt.Sprintf("%d:%s:%s", timestamp.UnixMilli(), transactionID.String(), key),
	))

	return hex.EncodeToString(hashSum[:])
}
