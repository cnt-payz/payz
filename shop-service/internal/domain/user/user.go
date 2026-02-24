package userdomain

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserID uuid.UUID

func NewUserID(userID string) (UserID, error) {
	u, err := uuid.Parse(userID)
	if err != nil {
		return UserID{}, status.Error(codes.InvalidArgument, "invalid user id")
	}

	return UserID(u), nil
}

func (u UserID) UUID() uuid.UUID {
	return uuid.UUID(u)
}

func (u UserID) String() string {
	return u.String()
}
