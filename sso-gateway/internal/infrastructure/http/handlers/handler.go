package handlers

import (
	"errors"
	"net/http"

	"github.com/cnt-payz/payz/sso-gateway/internal/application/services"
	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/http/middlewares"
	"github.com/cnt-payz/payz/sso-gateway/pkg/consts"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, service services.SSOGatewayService) (http.Handler, error) {
	if cfg == nil || service == nil {
		return nil, consts.ErrNilArgs
	}

	var router *gin.Engine
	switch cfg.Env {
	case "prod":
		router = gin.New()
		router.Use(gin.Recovery())
	case "dev", "local":
		router = gin.Default()
	default:
		return nil, errors.New("invalid environment")
	}
	router.Use(middlewares.RateLimitMiddleware(cfg.RateLimit.Limit, cfg.RateLimit.Window))

	routes := NewRoutes(service)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/login", routes.Login())
			v1.POST("/register", routes.Register())
			v1.POST("/refresh", routes.Refresh())
			v1.DELETE("/logout", routes.LogoutAll())

			v1.DELETE("/user", routes.LogoutAll())
			v1.GET("/user", routes.GetUser())
		}
	}

	return router, nil
}
