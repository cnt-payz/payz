package middlewareshttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("ac")
		if err != nil || token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		md := metadata.Pairs("authorization", "Bearer "+token)
		ctxWithAuth := metadata.NewOutgoingContext(ctx.Request.Context(), md)
		ctx.Request = ctx.Request.WithContext(ctxWithAuth)
		ctx.Next()
	}
}
