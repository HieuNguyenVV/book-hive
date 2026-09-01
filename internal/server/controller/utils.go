package controller

import (
	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/server/validator"
	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	_ = c.Error(err)

	appError, ok := err.(appErrors.AppError)
	if ok {
		c.AbortWithStatusJSON(appError.StatusCode, err)
		return
	}
	c.AbortWithStatusJSON(appErrors.ErrInternalServerError.StatusCode, err)
}

func handleBindError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	handleError(c, validator.BindError(err))
}
