package data

type (
	// User is the data type for user object
	User struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Password  string `json:"password" validate:"required"`
		Username  string `json:"username" validate:"required"`
		TokenHash string `json:"tokenhash"`
		CreatedAt string `json:"createdat"`
		UpdatedAt string `json:"updatedat"`
	}

	Balance struct {
		ID           int64   `json:"id"`
		UserID       string  `json:"userId"`
		IntegerPart  float64 `json:"integerPart"`
		FractionPart float64 `json:"fractionPart"`
		Currency     string  `json:"currency"`
	}
)
