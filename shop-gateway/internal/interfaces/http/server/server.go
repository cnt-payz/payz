package serverhttp

import (
	"fmt"
	"net/http"

	"github.com/cnt-payz/payz/shop-gateway/config"
	handlershttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/handlers"
	middlewareshttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/middlewares"
	routeshttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/routes"
	"github.com/gin-gonic/gin"
)

func NewServer(cfg *config.Config, handlersHTTP handlershttp.HTTPHandlers) *http.Server {
	r := gin.Default()
	r.Use(middlewareshttp.AuthMiddleware())
	routeshttp.Setup(r, handlersHTTP)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: r,
	}

	return srv
}
