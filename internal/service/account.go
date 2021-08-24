package service

import (
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"math"
	"my-bank-service/internal/config"
	"my-bank-service/internal/data"
	"my-bank-service/pkg/session"
)

var (
	SBP2RUB = 0.7523
	RUB     = "RUB"
	SBP     = "SBP"
)

type AccountService struct {
	session *session.Session
}

func NewAccountService(sf *session.SessionFactory) AccountInterface {
	return &AccountService{session: sf.GetSession()}
}

func (a *AccountService) AddFunds(sum float64) error {
	if sum == 0 {
		return nil
	}
	err := a.session.Begin()
	if err != nil {
		return err
	}
	defer a.session.Rollback()
	dialect := goqu.Dialect(config.Driver)
	dbBalance, err2 := a.getBalance(dialect)
	if err2 != nil {
		return err2
	}
	floatDbBalance := dbBalance.SBP + dbBalance.SPF
	floatDbBalance += sum

	sbp, _ := math.Modf(floatDbBalance)
	bal := data.SBPBalance{
		SBP: sbp,
		SPF: floatDbBalance - sbp,
	}
	err = a.upBalance(dialect, bal)
	if err != nil {
		return err
	}

	err = a.sumProfit()
	if err != nil {
		return err
	}

	err = a.session.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (a *AccountService) sumProfit() error {
	dialect := goqu.Dialect(config.Driver)
	dbBalance, err := a.getBalance(dialect)
	if err != nil {
		return err
	}
	floatDbBalance := dbBalance.SBP + dbBalance.SPF

	percent := (floatDbBalance * config.AddPercent) / 100
	floatDbBalance = floatDbBalance + percent

	fmt.Println(floatDbBalance)
	sbp, _ := math.Modf(floatDbBalance)
	bal := data.SBPBalance{
		SBP: sbp,
		SPF: floatDbBalance - sbp,
	}
	err = a.upBalance(dialect, bal)
	if err != nil {
		return err
	}
	return nil
}

func (a *AccountService) Withdraw(sum float64) error {
	if sum == 0 {
		return nil
	}
	err := a.session.Begin()
	if err != nil {
		return err
	}
	defer a.session.Rollback()
	dialect := goqu.Dialect(config.Driver)
	balance, err := a.getBalance(dialect)
	if err != nil {
		return err
	}
	floatBalance := balance.SBP + balance.SPF

	if floatBalance < sum {
		return errors.New("there are not enough funds in your account")
	}

	withdrawalAmount := floatBalance * config.MaxPercent / 100

	if sum > withdrawalAmount {
		return errors.New("the amount exceeds the bank limit")
	}

	floatBalance -= sum

	sbp, _ := math.Modf(floatBalance)
	bal := data.SBPBalance{
		SBP: sbp,
		SPF: floatBalance - sbp,
	}
	err = a.upBalance(dialect, bal)
	if err != nil {
		return err
	}
	err = a.session.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (a *AccountService) GetCurrency() (string, error) {

	dialect := goqu.Dialect(config.Driver)
	ds := dialect.From("account").Select("currency")
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		return "", err
	}
	res, err := a.session.Query(sqlStr)
	if err != nil {
		return "", err
	}
	defer res.Close()
	var currency string
	if res.Next() {
		err = res.Scan(&currency)
	}
	return currency, nil
}

func (a *AccountService) GetAccountCurrencyRate(cur string) (float64, error) {
	currency, err := a.GetCurrency()
	if err != nil {
		return 0, err
	}
	if cur == currency {
		return 1.0, nil
	}
	if currency == RUB && cur == SBP {
		return 1 / SBP2RUB, nil
	} else if currency == SBP && cur == RUB {
		return SBP2RUB, nil
	} else {
		return 0, errors.New("the bank does not support the specified currency")
	}
}

func (a *AccountService) GetBalance(cur string) (float64, error) {
	if cur == "" {
		cur = SBP
	}
	dialect := goqu.Dialect(config.Driver)
	balance, err := a.getBalance(dialect)
	if err != nil {
		return 0, err
	}
	var floatBalance float64
	switch cur {
	case RUB:
		floatBalance = (balance.SBP + balance.SPF) * SBP2RUB
	case SBP:
		floatBalance = balance.SBP + balance.SPF
	default:
		return 0, errors.New("the bank does not support the specified currency")
	}
	floatBalance = a.round(floatBalance, .0, 2)
	return floatBalance, nil

}

func (a *AccountService) getBalance(dialect goqu.DialectWrapper) (data.SBPBalance, error) {
	ds := dialect.From("account").Select("balance.sbp", "balance.spf").LeftJoin(
		goqu.T("balance"), goqu.On(goqu.Ex{"account.balanceID": goqu.I("balance.id")}))
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		return data.SBPBalance{}, err
	}
	res, err := a.session.Query(sqlStr)
	if err != nil {
		return data.SBPBalance{}, err
	}
	defer res.Close()
	var dbBalance data.SBPBalance
	if res.Next() {
		err = res.Scan(&dbBalance.SBP, &dbBalance.SPF)
	}
	return dbBalance, nil
}

func (a *AccountService) upBalance(dialect goqu.DialectWrapper, balance data.SBPBalance) error {
	ds := dialect.From("account").Select("balanceId")
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		return err
	}
	res, err := a.session.Query(sqlStr)
	if err != nil {
		return err
	}
	defer res.Close()
	var balanceId int64
	if res.Next() {
		err = res.Scan(&balanceId)
		if err != nil {
			return err
		}
	}
	upDs := dialect.Update("balance").Set(
		goqu.Record{"sbp": balance.SBP, "spf": balance.SPF}).Where(goqu.Ex{"id": balanceId})

	newSqlStr, _, err := upDs.ToSQL()
	if err != nil {
		return err
	}
	_, err = a.session.Exec(newSqlStr)
	if err != nil {
		return err
	}
	return nil
}

func (a *AccountService) round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
