package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"my-bank-service/internal/config"
	data2 "my-bank-service/internal/data"
	"my-bank-service/internal/reposytory"
	"my-bank-service/internal/service"
	utils2 "my-bank-service/internal/utils"
	"my-bank-service/internal/validation"
	"my-bank-service/pkg/logging"
	"net/http"
	"strings"
)

var (
	ErrUserAlreadyExists = fmt.Sprintf("User already exists with the given email")
	ErrUserNotFound      = fmt.Sprintf("No user account exists with given email. Please sign in first")
	UserCreationFailed   = fmt.Sprintf("Unable to create user.Please try again later")

	PgDuplicateKeyMsg = "duplicate key value violates unique constraint"
	PgNoRowsMsg       = "no rows in result set"
)

// UserKey is used as a key for storing the User object in context at middleware
type UserKey struct{}

// UserIDKey is used as a key for storing the UserID in context at middleware
type UserIDKey struct{}

// VerificationDataKey is used as the key for storing the VerificationData in context at middleware
type VerificationDataKey struct{}

// AuthHandler wraps instances needed to perform operations on user object
type AuthHandler struct {
	logger      logging.Logger
	configs     *config.Configurations
	validator   *validation.Validation
	repo        reposytory.UserRepository
	authService service.Authentication
}

// NewAuthHandler returns a new AuthHandler instance
func NewAuthHandler(l logging.Logger, c *config.Configurations, v *validation.Validation, r reposytory.UserRepository, auth service.Authentication) *AuthHandler {
	return &AuthHandler{
		logger:      l,
		configs:     c,
		validator:   v,
		repo:        r,
		authService: auth,
	}
}

func (ah *AuthHandler) Routes(engine *gin.Engine) {
	user := engine.Group(config.GroupPath)
	{
		user.Use(ah.MiddlewareValidateUser)
		user.POST(config.LoginPath, ah.Login)
		user.POST(config.SignUpPath, ah.Signup)
		user.Use(ah.MiddlewareValidateAccessToken)
	}

	refToken := engine.Group(config.RefreshTokenPath)
	{
		refToken.Use(ah.MiddlewareValidateRefreshToken)
		refToken.GET("", ah.RefreshToken)
	}

}

// GenericResponse is the format of our response
type GenericResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ValidationError is a collection of validation error messages
type ValidationError struct {
	Errors []string `json:"errors"`
}

// TokenResponse below data types are used for encoding and decoding b/t go types and json
type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type AuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	Username     string `json:"username"`
}

type UsernameUpdate struct {
	Username string `json:"username"`
}

type CodeVerificationReq struct {
	Code string `json:"code"`
	Type string `json:"type"`
}

type PasswordResetReq struct {
	NewPassword   string `json:"new_password"`
	NewPasswordRe string `json:"new_password_re"`
	OldPassword   string `json:"old_password"`
}

// Login handles login request
// @Summary Login
// @Tags auth
// @Description user login
// @ID user-login
// @Accept json
// @Produce json
// @Param input body swagger.UserLogin true "username and password"
// @Success 200 {integer} integer 1
// @Failure 400,404,500 {integer} integer 2
// @Router /login/ [post]
func (ah *AuthHandler) Login(ctx *gin.Context) {

	ctx.Set("Content-Type", "application/json")

	reqUser := ctx.Request.Context().Value(UserKey{}).(data2.User)
	user, err := ah.repo.GetUserByUserName(reqUser.Username)
	if err != nil {
		ah.logger.Error("error fetching the user", "error", err)
		errMsg := err.Error()
		if strings.Contains(errMsg, PgNoRowsMsg) {
			ctx.AbortWithStatus(http.StatusBadRequest)
			_ = data2.ToJSON(&GenericResponse{Status: false, Message: ErrUserNotFound}, ctx.Writer)
		} else {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			_ = data2.ToJSON(&GenericResponse{Status: false, Message: "Unable to retrieve user from database.Please try again later"}, ctx.Writer)
		}
		return
	}

	if valid := ah.authService.Authenticate(&reqUser, user); !valid {
		ah.logger.Debug("Authetication of user failed")
		ctx.AbortWithStatus(http.StatusBadRequest)
		_ = data2.ToJSON(&GenericResponse{Status: false, Message: "Incorrect username or password"}, ctx.Writer)
		return
	}

	accessToken, err := ah.authService.GenerateAccessToken(user)
	if err != nil {
		ah.logger.Error("unable to generate access token", "error", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		_ = data2.ToJSON(&GenericResponse{Status: false, Message: "Unable to login the user. Please try again later"}, ctx.Writer)
		return
	}
	refreshToken, err := ah.authService.GenerateRefreshToken(user)
	if err != nil {
		ah.logger.Error("unable to generate refresh token", "error", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		_ = data2.ToJSON(&GenericResponse{Status: false, Message: "Unable to login the user. Please try again later"}, ctx.Writer)
		return
	}

	ah.logger.Debug("successfully generated tokens")
	ctx.AbortWithStatus(http.StatusOK)
	_ = data2.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Successfully logged in",
		Data:    &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, Username: user.Username},
	}, ctx.Writer)
}

// Signup handles signup request
// @Summary Signup
// @Tags auth
// @Description create user
// @ID create-user
// @Accept json
// @Produce json
// @Param input body swagger.UserSignUp true "signup user"
// @Success 201 {integer} integer 1
// @Failure 400,500 {integer} integer 2
// @Router /user/add/ [post]
func (ah *AuthHandler) Signup(ctx *gin.Context) {
	ctx.Set("Content-Type", "application/json")

	user := ctx.Request.Context().Value(UserKey{}).(data2.User)
	hashedPass, err := ah.hashPassword(user.Password)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		_ = data2.ToJSON(&GenericResponse{Status: false, Message: UserCreationFailed}, ctx.Writer)
		return
	}
	user.Password = hashedPass
	user.TokenHash = utils2.GenerateRandomString(15)

	err = ah.repo.Create(&user)
	if err != nil {
		ah.logger.Error("unable to insert user to database", "error", err)
		errMsg := err.Error()
		if strings.Contains(errMsg, PgDuplicateKeyMsg) {
			ctx.AbortWithStatus(http.StatusBadRequest)
			_ = data2.ToJSON(&GenericResponse{Status: false, Message: ErrUserAlreadyExists}, ctx.Writer)
		} else {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			_ = data2.ToJSON(&GenericResponse{Status: false, Message: UserCreationFailed}, ctx.Writer)
			return
		}
		return
	}

	ah.logger.Debug("user created successfully")
	ctx.AbortWithStatus(http.StatusCreated)
	_ = data2.ToJSON(&GenericResponse{Status: true, Message: "User created successfully."}, ctx.Writer)
}

// RefreshToken handles refresh token request
// @Summary RefreshToken
// @Security ApiKeyAuth
// @Tags auth
// @Description refresh token
// @ID ref-token
// @Accept json
// @Produce json
// @Success 200 {integer} integer 1
// @Failure 500 {integer} integer 2
// @Router /refresh-token/ [get]
func (ah *AuthHandler) RefreshToken(ctx *gin.Context) {
	ctx.Set("Content-Type", "application/json")

	user := ctx.Request.Context().Value(UserKey{}).(data2.User)
	accessToken, err := ah.authService.GenerateAccessToken(&user)
	if err != nil {
		ah.logger.Error("unable to generate access token", "error", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		_ = data2.ToJSON(&GenericResponse{Status: false, Message: "Unable to generate access token.Please try again later"}, ctx.Writer)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
	_ = data2.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Successfully generated new access token",
		Data:    &TokenResponse{AccessToken: accessToken},
	}, ctx.Writer)
}

// hashPassword generates an encrypted password
func (ah *AuthHandler) hashPassword(password string) (string, error) {

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		ah.logger.Error("unable to hash password", "error", err)
		return "", err
	}

	return string(hashedPass), nil
}
