package filemanager

import (
	"fmt"
	"os"
	"strings"

	"github.com/cnt-payz/payz/sso-service/pkg/consts"
)

type FileManager struct {
}

func New() *FileManager {
	return &FileManager{}
}

func (fm *FileManager) LoadFile(path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, consts.ErrInvalidPath
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return bytes, nil
}
