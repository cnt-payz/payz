package handlershttp

import (
	"errors"
	"log/slog"
	"net/http"

	shopusecase "github.com/cnt-payz/payz/shop-gateway/internal/application/usecase/shop"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HTTPHandlers interface {
	CreateShop() gin.HandlerFunc
	GetShopByOwner() gin.HandlerFunc
	GetShopsByUserID() gin.HandlerFunc
	UpdateShopByID() gin.HandlerFunc
	DeleteShopByID() gin.HandlerFunc
}

type httpHandlers struct {
	shopUsecase shopusecase.ShopUsecase
	log         *slog.Logger
}

func NewHTTPHandlers(shopUsecase shopusecase.ShopUsecase, log *slog.Logger) HTTPHandlers {
	return &httpHandlers{
		shopUsecase: shopUsecase,
		log:         log,
	}
}

func (hh *httpHandlers) CreateShop() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hh.log.Debug("received request to create shop")

		var req struct {
			Name string `json:"name"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &gin.H{
				"error": "failed to parse name",
			})
			return
		}

		shop, err := hh.shopUsecase.CreateShop(ctx.Request.Context(), req.Name)
		if err != nil {
			hh.handleError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, &gin.H{
			"shop": *shop,
		})
	}
}

func (hh *httpHandlers) GetShopByOwner() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hh.log.Debug("received request to get shop")

		id := ctx.Query("id")

		shop, err := hh.shopUsecase.GetShopByOwner(ctx.Request.Context(), id)
		if err != nil {
			hh.handleError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, &gin.H{
			"shop": *shop,
		})
	}
}

func (hh *httpHandlers) GetShopsByUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hh.log.Debug("received request to get shops")

		shops, err := hh.shopUsecase.GetShopsByUserID(ctx.Request.Context())
		if err != nil {
			hh.handleError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, &gin.H{
			"shops": shops,
		})
	}
}

func (hh *httpHandlers) UpdateShopByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hh.log.Debug("received request to update shop")

		var req struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &gin.H{
				"error": "failed to parse id, name",
			})
			return
		}

		shop, err := hh.shopUsecase.UpdateShopByID(ctx.Request.Context(), req.ID, req.Name)
		if err != nil {
			hh.handleError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, &gin.H{
			"shop": *shop,
		})
	}
}

func (hh *httpHandlers) DeleteShopByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hh.log.Debug("received request to delete shop")

		var req struct {
			ID string `json:"id"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &gin.H{
				"error": "failed to parse id",
			})
			return
		}

		err := hh.shopUsecase.DeleteShopByID(ctx.Request.Context(), req.ID)
		if err != nil {
			hh.handleError(ctx, err)
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (hh *httpHandlers) handleError(ctx *gin.Context, err error) {
	if errors.Is(err, shopusecase.ErrEmptyID) || errors.Is(err, shopusecase.ErrEmptyName) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": st.Message()})
		case codes.NotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": st.Message()})
		case codes.Unauthenticated:
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
		case codes.FailedPrecondition:
			ctx.AbortWithStatusJSON(http.StatusPreconditionFailed, gin.H{"error": st.Message()})
		default:
			hh.log.Error("handler error", slog.String("error", err.Error()))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
}
