package handlers

import (
	"net/http"
	"strings"

	"github.com/cnt-payz/payz/sso-gateway/internal/application/dtos"
	"github.com/cnt-payz/payz/sso-gateway/internal/application/services"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	service services.SSOGatewayService
}

func NewRoutes(service services.SSOGatewayService) *Routes {
	return &Routes{
		service: service,
	}
}

func (r *Routes) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.LoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Login(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Register() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.RegisterRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Register(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Refresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.RefreshRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Refresh(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) LogoutAll() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if code, err := r.service.LogoutAll(ctx); err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func (r *Routes) GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			resp *dtos.User
			code int
			err  error
		)

		if id := strings.TrimSpace(ctx.Query("id")); id != "" {
			resp, code, err = r.service.GetUserByID(ctx, id)
		} else if email := strings.TrimSpace(ctx.Query("email")); email != "" {
			resp, code, err = r.service.GetUserByEmail(ctx, email)
		} else {
			resp, code, err = r.service.GetSelfUser(ctx)
		}

		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}
