package rest

import (
	"errors"
	"net/http"

	"project-template/internal/usecase/checkout"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	applogger "project-template/pkg/logger"
)

type CheckoutHandler struct {
	service *checkout.Service
	log     *applogger.Logger
}

func NewCheckoutHandler(service *checkout.Service, logger *applogger.Logger) *CheckoutHandler {
	if logger == nil {
		logger = applogger.Wrap(nil)
	}
	return &CheckoutHandler{service: service, log: logger.Named("checkout-handler")}
}

func (h *CheckoutHandler) Checkout(c *gin.Context) {
	log := h.log
	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnCtx(c.Request.Context(), "checkout request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "checkout request", zap.String("user_id", req.UserID), zap.Int("item_count", len(req.Items)))
	result, err := h.service.Checkout(c.Request.Context(), req.ToInput())
	if err != nil {
		log.WarnCtx(c.Request.Context(), "checkout request failed", zap.String("user_id", req.UserID), zap.Error(err))
		switch {
		case errors.Is(err, checkout.ErrUserNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, checkout.ErrInvalidCheckout):
			respondError(c, http.StatusBadRequest, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "checkout request succeeded", zap.String("user_id", req.UserID), zap.String("order_id", result.Order.ID))
	respondJSON(c, http.StatusCreated, NewCheckoutResponse(result))
}
