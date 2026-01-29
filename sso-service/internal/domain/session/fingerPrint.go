package session

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/cnt-payz/payz/sso-service/pkg/consts"
)

type FingerPrint string

func NewFingerPrint(clientIP, userAgent string) (FingerPrint, error) {
	if len(userAgent) < 16 {
		return "", consts.ErrInvalidUserAgent
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", clientIP, userAgent[:16])))
	return FingerPrint(hex.EncodeToString(hash[:])), nil
}

func (fp FingerPrint) Value() string {
	return string(fp)
}
