package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"my-bank-service/internal/data"
	"net/http"
	"strings"
)

// MiddlewareValidateUser validates the user in the request
func (ah *AuthHandler) MiddlewareValidateUser(ctx *gin.Context) {
	ctx.Set("Content-Type", "application/json")

	ah.logger.Debug("user json received")

	user := &data.User{}
	err := ctx.ShouldBindBodyWith(user, binding.JSON)
	if err != nil {
		ah.logger.Error("deserialization of user json failed", "error", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: err.Error()}, ctx.Writer)
		return
	}

	// validate the user
	errs := ah.validator.Validate(user)
	if len(errs) != 0 {
		ah.logger.Error("validation of user json failed", "error", errs)
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: strings.Join(errs.Errors(), ",")}, ctx.Writer)
		return
	}

	// add the user to the context
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), UserKey{}, *user))

	// call the next handler
	ctx.Next()

}

// MiddlewareValidateAccessToken validates whether the request contains a bearer token
// it also decodes and authenticates the given token
func (ah *AuthHandler) MiddlewareValidateAccessToken(ctx *gin.Context) {
	ctx.Set("Content-Type", "application/json")
	ah.logger.Debug("validating access token")
	token, err := extractToken(ctx)
	if err != nil {
		ah.logger.Error("token not provided or malformed")
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Authentication failed. Token not provided or malformed"}, ctx.Writer)
		return
	}
	ah.logger.Debug("token present in header")

	userID, err := ah.authService.ValidateAccessToken(token)
	if err != nil {
		ah.logger.Error("token validation failed", "error", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Authentication failed. Invalid token"}, ctx.Writer)
		return
	}
	ah.logger.Debug("access token validated")

	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), UserIDKey{}, userID))

	// call the next handler
	ctx.Next()

}

// MiddlewareValidateRefreshToken validates whether the request contains a bearer token
// it also decodes and authenticates the given token
func (ah *AuthHandler) MiddlewareValidateRefreshToken(ctx *gin.Context) {

	ctx.Set("Content-Type", "application/json")

	ah.logger.Debug("validating refresh token")
	ah.logger.Debug("auth header", ctx.Request.Header.Get("Authorization"))
	token, err := extractToken(ctx)
	if err != nil {
		ah.logger.Error("token not provided or malformed")
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Authentication failed. Token not provided or malformed"}, ctx.Writer)
		return
	}
	ah.logger.Debug("token present in header", token)

	userID, customKey, err := ah.authService.ValidateRefreshToken(token)
	if err != nil {
		ah.logger.Error("token validation failed", "error", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Authentication failed. Invalid token"}, ctx.Writer)
		return
	}
	ah.logger.Debug("refresh token validated")

	user, err := ah.repo.GetUserByID(userID)
	if err != nil {
		ah.logger.Error("invalid token: wrong userID while parsing", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Unable to fetch corresponding user"}, ctx.Writer)
		return
	}

	actualCustomKey := ah.authService.GenerateCustomKey(user.ID, user.TokenHash)
	if customKey != actualCustomKey {
		ah.logger.Debug("wrong token: authetincation failed")
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data.ToJSON(&GenericResponse{Status: false, Message: "Authentication failed. Invalid token"}, ctx.Writer)
		return
	}

	// add the user to the context
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), UserKey{}, *user))

	// call the next handler
	ctx.Next()
}

// extractToken checks for the presence of authorization headers
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Token not provided or malformed")
	}
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", errors.New("Token not provided or malformed")
	}
	return headerParts[1], nil
}
