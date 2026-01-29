package user_test

import (
	"errors"
	"testing"

	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
)

type input struct {
	email, password string
}

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   input
		WantErr error
	}{
		{
			Name: "base",
			Input: input{
				email:    "mail@example.com",
				password: "very_secret_password",
			},
			WantErr: nil,
		},
		{
			Name: "invalid_mail",
			Input: input{
				email:    "mail@example.",
				password: "very_secret_password",
			},
			WantErr: consts.ErrInvalidEmail,
		},
		{
			Name: "invalid_mail",
			Input: input{
				email:    "mail@.com",
				password: "very_secret_password",
			},
			WantErr: consts.ErrInvalidEmail,
		},
		{
			Name: "invalid_mail",
			Input: input{
				email:    "",
				password: "very_secret_password",
			},
			WantErr: consts.ErrInvalidEmail,
		},
		{
			Name: "invalid_password",
			Input: input{
				email:    "mail@example.com",
				password: "",
			},
			WantErr: consts.ErrInvalidPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			_, err := user.NewEmail(tc.Input.email)
			if err != nil && !errors.Is(err, tc.WantErr) {
				t.Errorf("want %v, got %v", tc.WantErr, err)
			}

			_, err = user.NewPasswordHash(tc.Input.password)
			if err != nil && !errors.Is(err, tc.WantErr) {
				t.Errorf("want %v, got %v", tc.WantErr, err)
			}
		})
	}
}
