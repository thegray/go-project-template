package rest

import (
	"errors"
	"net/http"

	"auth-service/internal/user"
	applogger "auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *user.Service
	log     *applogger.Logger
}

func NewUserHandler(service *user.Service, logger *applogger.Logger) *UserHandler {
	if logger == nil {
		logger = applogger.Wrap(nil)
	}
	return &UserHandler{service: service, log: logger}
}

func (h *UserHandler) Login(c *gin.Context) {
	log := h.log
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnCtx(c.Request.Context(), "login request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "login request", zap.String("email", req.Email))
	result, err := h.service.Login(c.Request.Context(), user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.WarnCtx(c.Request.Context(), "login request failed", zap.String("email", req.Email), zap.Error(err))
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			respondError(c, http.StatusUnauthorized, err.Error())
		case errors.Is(err, user.ErrNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "login request succeeded", zap.String("user_id", result.ID))
	respondJSON(c, http.StatusOK, NewUserResponse(result))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	log := h.log
	var param UserPathParam
	if err := c.ShouldBindUri(&param); err != nil {
		log.WarnCtx(c.Request.Context(), "get user request invalid", zap.Error(err))
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.InfoCtx(c.Request.Context(), "get user request", zap.String("user_id", param.ID))
	result, err := h.service.GetByID(c.Request.Context(), param.ID)
	if err != nil {
		log.WarnCtx(c.Request.Context(), "get user request failed", zap.String("user_id", param.ID), zap.Error(err))
		switch {
		case errors.Is(err, user.ErrNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.InfoCtx(c.Request.Context(), "get user request succeeded", zap.String("user_id", result.ID))
	respondJSON(c, http.StatusOK, NewUserResponse(result))
}
