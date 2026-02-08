package shopdomain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewShop(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		Name     string
		ShopName string
		Err      error
	}{
		{
			Name:     "ok",
			ShopName: "name",
			Err:      nil,
		}, {
			Name:     "empty name",
			ShopName: "   ",
			Err:      ErrEmptyShopName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := NewShop(id, tt.ShopName)
			if !errors.Is(err, tt.Err) {
				t.Errorf("expected: %v have %v", tt.Err, err)
			}
		})
	}
}

func TestNewName(t *testing.T) {
	tests := []struct {
		Name    string
		NewName string
		Err     error
	}{
		{
			Name:    "ok",
			NewName: "new name",
			Err:     nil,
		}, {
			Name:    "empty name",
			NewName: "  ",
			Err:     ErrEmptyShopName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := NewName(tt.NewName)
			if !errors.Is(err, tt.Err) {
				t.Errorf("expected: %v have: %v", tt.Err, err)
			}
		})
	}
}
