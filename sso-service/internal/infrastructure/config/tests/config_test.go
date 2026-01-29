package config_test

import (
	"testing"

	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		Name      string
		InputPath string
		WantErr   bool
	}{
		{Name: "base", InputPath: "./test.yml", WantErr: false},
		{Name: "invalid_path", InputPath: "./test.", WantErr: true},
		{Name: "empty_path", InputPath: "", WantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := config.New(tc.InputPath); err != nil && !tc.WantErr {
				t.Errorf("want %v got %v", tc.WantErr, err != nil)
			}
		})
	}
}
