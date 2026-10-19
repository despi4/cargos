package http

import (
	"errors"
	"net/http"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/despi4/cargos/auth_service/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc ports.AuthService
}

func NewAuthHandler(svc ports.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	user, err := h.authSvc.Register(c, input.Username, input.Email, input.Password)
	if err != nil {
		statusCode := http.StatusInternalServerError

		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			statusCode = http.StatusConflict
		case errors.Is(err, domain.ErrPasswordTooShort),
			errors.Is(err, domain.ErrPasswordTooLong),
			errors.Is(err, domain.ErrPasswordNoDigit),
			errors.Is(err, domain.ErrPasswordNoLower),
			errors.Is(err, domain.ErrPasswordNoUpper),
			errors.Is(err, domain.ErrPasswordNoSpecial),
			errors.Is(err, domain.ErrEmailInvalidFormat),
			errors.Is(err, domain.ErrUsernameTooShort),
			errors.Is(err, domain.ErrUsernameTooLong):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, Response{
			Success: false,
			Error:   err.Error(),
			Code:    statusCode,
		})
		return
	}

	res := Response{
		Success: true,
		Data:    user.Email,
		Message: "New User Created",
		Code:    http.StatusCreated,
	}

	c.JSON(201, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	tokenPair, err := h.authSvc.Login(c, input.Email, input.Password, input.ClientIP, input.UserAgent)
	if err != nil {
		statusCode := http.StatusInternalServerError

		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			statusCode = http.StatusConflict
		case errors.Is(err, domain.ErrPasswordTooShort),
			errors.Is(err, domain.ErrPasswordTooLong),
			errors.Is(err, domain.ErrPasswordNoDigit),
			errors.Is(err, domain.ErrPasswordNoLower),
			errors.Is(err, domain.ErrPasswordNoUpper),
			errors.Is(err, domain.ErrPasswordNoSpecial),
			errors.Is(err, domain.ErrEmailInvalidFormat),
			errors.Is(err, domain.ErrUsernameTooShort),
			errors.Is(err, domain.ErrUsernameTooLong):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, Response{
			Success: false,
			Error:   err.Error(),
			Code:    statusCode,
		})
		return
	}

	// 1. Устанавливаем Refresh-токен в защищенную куку на 7 дней (7 * 24 * 3600 сек)
	c.SetCookie(
		"refresh_token",        // имя куки
		tokenPair.RefreshToken, // значение
		7*24*3600,              // maxAge в секундах
		"/api/v1/auth",         // path: кука будет отправляться ТОЛЬКО на роуты авторизации!
		"",                     // domain: текущий домен
		false,                  // secure: true в проде (HTTPS), false для localhost
		true,                   // httpOnly: true (JS не имеет доступа!)
	)

	// 2. В JSON отдаем только AccessToken
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"access_token": tokenPair.AccessToken,
			"token_type":   "Bearer",
			"expires_in":   900,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	h.authSvc.RefreshToken()
}
