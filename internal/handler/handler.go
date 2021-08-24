package handler

import (
	"github.com/gin-gonic/gin"
	"my-bank-service/internal/data"
	"my-bank-service/internal/service"
	"my-bank-service/pkg/logging"
	"my-bank-service/pkg/session"
	"net/http"
)

type RouteDivider interface {
	Routes(*gin.Engine)
}

type accountHandler struct {
	service service.AccountInterface
	logger  logging.Logger
}

func NewAccountHandler(l logging.Logger, sf *session.SessionFactory) RouteDivider {
	return &accountHandler{logger: l, service: service.NewAccountService(sf)}
}

func (a *accountHandler) Routes(route *gin.Engine) {
	bank := route.Group("account")
	{
		bank.GET("/currency/", a.GetAccountCurrency)
		bank.POST("/addFund/", a.AddFund)
		bank.POST("/balance/", a.GetAccountBalance)
		bank.POST("/withDraw/", a.WithdrawMoney)
		bank.POST("/currencyRate/", a.GetCurrencyRate)
	}
}

func (a *accountHandler) AddFund(ctx *gin.Context) {
	var addHeader data.FundHeader
	if err := ctx.BindJSON(&addHeader); err != nil {
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
	}
	if addHeader.Currency == service.RUB {
		addHeader.Balance /= service.SBP2RUB
	}
	err := a.service.AddFunds(addHeader.Balance)
	if err != nil {
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data.GenericResponse{Status: true, Message: "amount added successfully"})
}

func (a *accountHandler) GetAccountCurrency(ctx *gin.Context) {
	currency, err := a.service.GetCurrency()
	if err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}
	currStruct := struct {
		Currency string `json:"currency"`
	}{currency}
	ctx.JSON(http.StatusOK, data.GenericResponse{Status: true, Message: "", Data: &currStruct})
}

func (a *accountHandler) GetAccountBalance(ctx *gin.Context) {
	currStruct := struct {
		Currency string `json:"currency"`
	}{}
	if err := ctx.BindJSON(&currStruct); err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}

	balance, err := a.service.GetBalance(currStruct.Currency)
	if err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, data.GenericResponse{Message: err.Error()})
		return
	}
	balanceStruct := struct {
		Balance float64 `json:"balance"`
	}{balance}
	ctx.JSON(http.StatusOK, data.GenericResponse{Status: true, Message: "", Data: &balanceStruct})
}

func (a *accountHandler) GetCurrencyRate(ctx *gin.Context) {
	currStruct := struct {
		Currency string `json:"currency"`
	}{}
	if err := ctx.BindJSON(&currStruct); err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}
	rate, err := a.service.GetAccountCurrencyRate(currStruct.Currency)
	if err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, data.GenericResponse{Message: err.Error()})
		return
	}
	rateStruct := struct {
		Rate float64 `json:"rate"`
	}{rate}
	ctx.JSON(http.StatusOK, data.GenericResponse{Status: true, Message: "", Data: &rateStruct})
}

func (a *accountHandler) WithdrawMoney(ctx *gin.Context) {
	var whitDrawHeader data.FundHeader
	if err := ctx.BindJSON(&whitDrawHeader); err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}
	if whitDrawHeader.Currency == service.RUB {
		whitDrawHeader.Balance /= service.SBP2RUB
	}
	err := a.service.Withdraw(whitDrawHeader.Balance)
	if err != nil {
		a.logger.Error(err)
		ctx.JSON(http.StatusNotFound, data.GenericResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data.GenericResponse{Status: true, Message: "money was withdrawn from the account"})
}
