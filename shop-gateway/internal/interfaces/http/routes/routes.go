package routeshttp

import (
	handlershttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, httpHandlers handlershttp.HTTPHandlers) {
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/shop", httpHandlers.CreateShop())
			v1.GET("/shop", httpHandlers.GetShopByOwner())
			v1.GET("/shops", httpHandlers.GetShopsByUserID())
			v1.PATCH("/shop", httpHandlers.UpdateShopByID())
			v1.DELETE("/shop", httpHandlers.DeleteShopByID())
		}
	}
}
