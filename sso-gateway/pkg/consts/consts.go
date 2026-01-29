package consts

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	CodeMap = map[codes.Code]int{
		codes.InvalidArgument:  http.StatusBadRequest,
		codes.Canceled:         http.StatusRequestTimeout,
		codes.DeadlineExceeded: http.StatusRequestTimeout,
		codes.NotFound:         http.StatusNotFound,
		codes.Internal:         http.StatusInternalServerError,
		codes.AlreadyExists:    http.StatusConflict,
		codes.PermissionDenied: http.StatusForbidden,
		codes.Unauthenticated:  http.StatusUnauthorized,
	}
)
