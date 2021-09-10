package swagger

type (
	UserSignUp struct {
		Email    string `json:"email" validate:"required"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	UserLogin struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
)
