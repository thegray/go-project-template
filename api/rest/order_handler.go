package rest

import (
	"errors"
	"net/http"

	"project-template/internal/order"
	"project-template/internal/shared"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	applogger "project-template/pkg/logger"
)

type OrderHandler struct {
	service *order.Service
	log     *applogger.Logger
}

func NewOrderHandler(service *order.Service, logger *applogger.Logger) *OrderHandler {
	if logger == nil {
		logger = applogger.Wrap(nil)
	}
	return &OrderHandler{service: service, log: logger.Named("order-handler")}
}

func (h *OrderHandler) Create(c *gin.Context) {
	log := h.log
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnCtx(c.Request.Context(), "create order request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "create order request", zap.String("user_id", req.UserID), zap.Int("item_count", len(req.Items)), zap.Int("discount_percent", req.DiscountPercent))
	result, err := h.service.Create(c.Request.Context(), req.ToCreateInput())
	if err != nil {
		log.WarnCtx(c.Request.Context(), "create order request failed", zap.String("user_id", req.UserID), zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "create order request succeeded", zap.String("order_id", result.ID))
	respondJSON(c, http.StatusCreated, NewOrderResponse(result))
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	log := h.log
	var param OrderPathParam
	if err := c.ShouldBindUri(&param); err != nil {
		log.WarnCtx(c.Request.Context(), "get order request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "get order request", zap.String("order_id", param.ID))
	result, err := h.service.GetByID(c.Request.Context(), param.ID)
	if err != nil {
		log.WarnCtx(c.Request.Context(), "get order request failed", zap.String("order_id", param.ID), zap.Error(err))
		switch {
		case errors.Is(err, order.ErrNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "get order request succeeded", zap.String("order_id", result.ID))
	respondJSON(c, http.StatusOK, NewOrderResponse(result))
}

func (h *OrderHandler) List(c *gin.Context) {
	log := h.log
	var query OrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.WarnCtx(c.Request.Context(), "list orders request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	log.InfoCtx(c.Request.Context(), "list orders request", zap.String("user_id", query.UserID), zap.Int("page", query.Page), zap.Int("limit", query.Limit))
	result, err := h.service.List(c.Request.Context(), query.UserID, shared.Pagination{
		Page:  query.Page,
		Limit: query.Limit,
	})
	if err != nil {
		log.WarnCtx(c.Request.Context(), "list orders request failed", zap.String("user_id", query.UserID), zap.Error(err))
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "list orders request succeeded", zap.String("user_id", query.UserID), zap.Int("count", len(result)))
	respondJSON(c, http.StatusOK, NewOrderResponses(result))
}

func (h *OrderHandler) Update(c *gin.Context) {
	log := h.log
	var param OrderPathParam
	if err := c.ShouldBindUri(&param); err != nil {
		log.WarnCtx(c.Request.Context(), "update order request invalid: path", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnCtx(c.Request.Context(), "update order request invalid: body", zap.String("order_id", param.ID), zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "update order request", zap.String("order_id", param.ID), zap.String("user_id", req.UserID), zap.Int("item_count", len(req.Items)), zap.Int("discount_percent", req.DiscountPercent))
	result, err := h.service.Update(c.Request.Context(), param.ID, req.ToUpdateInput())
	if err != nil {
		log.WarnCtx(c.Request.Context(), "update order request failed", zap.String("order_id", param.ID), zap.Error(err))
		switch {
		case errors.Is(err, order.ErrNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "update order request succeeded", zap.String("order_id", result.ID))
	respondJSON(c, http.StatusOK, NewOrderResponse(result))
}

func (h *OrderHandler) Delete(c *gin.Context) {
	log := h.log
	var param OrderPathParam
	if err := c.ShouldBindUri(&param); err != nil {
		log.WarnCtx(c.Request.Context(), "delete order request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "delete order request", zap.String("order_id", param.ID))
	if err := h.service.Delete(c.Request.Context(), param.ID); err != nil {
		log.WarnCtx(c.Request.Context(), "delete order request failed", zap.String("order_id", param.ID), zap.Error(err))
		switch {
		case errors.Is(err, order.ErrNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "delete order request succeeded", zap.String("order_id", param.ID))
	c.Status(http.StatusNoContent)
}
