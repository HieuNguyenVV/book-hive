package middleware

import (
	"net/http"
	"strings"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/jwt"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	RefreshTokenHeader  = "RefreshToken"
	BearerPrefix        = "Bearer"
	CtxUserKey          = "user"
	CtxRefreshTokenKey  = "refresh_token"
)

func NewAuthMiddleware(logger log.Logger, userService service.UserService, jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(AuthorizationHeader)
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		token = strings.TrimPrefix(token, BearerPrefix+" ")
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			handleAuthError(c, err)
			return
		}
		if claims == nil {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		if claims.Subject == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		user, err := userService.GetUserByID(c.Request.Context(), claims.Subject)
		if err != nil {
			handleAuthError(c, err)
			return
		}
		if user.Status != model.UserStatusActive {
			handleAuthError(c, appErrors.ErrUnauthorized)
			return
		}
		c.Set(CtxUserKey, user)
		c.Next()
	}
}

func handleAuthError(ctx *gin.Context, err error) {
	handleError(ctx, err, appErrors.ErrUnauthorized)
}

func HandleWSAuthError(ctx *gin.Context, err error) {
	handleError(ctx, err, appErrors.ErrUnauthorized)
}

func handleError(ctx *gin.Context, err error, defaultAppError appErrors.AppError) {
	if err == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, defaultAppError)
		return
	}
	_ = ctx.Error(err)
	if appError, ok := err.(appErrors.AppError); ok {
		ctx.AbortWithStatusJSON(appError.StatusCode, err)
	} else {
		err := defaultAppError.Wrap(err)
		ctx.AbortWithStatusJSON(err.StatusCode, err)
	}
}

func NewRefreshTokenMiddleware(logger log.Logger, userService service.UserService, jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(RefreshTokenHeader)
		if token == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		userID, err := jwtService.ValidateRefreshToken(token)
		if err != nil {
			handleAuthError(c, err)
			return
		}

		if userID == "" {
			handleAuthError(c, appErrors.ErrInvalidToken)
			return
		}
		user, err := userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			handleAuthError(c, err)
			return
		}
		if user.Status != model.UserStatusActive {
			handleAuthError(c, appErrors.ErrUnauthorized)
			return
		}
		c.Set(CtxRefreshTokenKey, token)
		c.Next()
	}
}
