package data

type (
	GenericResponse struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	SBPBalance struct {
		SBP float64 `json:"sbp"`
		SPF float64 `json:"spf"`
	}

	Account struct {
		Id       int64      `json:"id"`
		Balance  SBPBalance `json:"balance"`
		Currency string     `json:"currency"`
	}

	FundHeader struct {
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}
)
