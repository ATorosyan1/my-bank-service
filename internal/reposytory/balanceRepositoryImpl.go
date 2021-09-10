package reposytory

import (
	"github.com/doug-martin/goqu/v9"
	"my-bank-service/internal/config"
	"my-bank-service/internal/data"
	"my-bank-service/pkg/logging"
	"my-bank-service/pkg/session"
)

// balanceRepository has the implementation of the db methods.
type balanceRepository struct {
	session *session.Session
	logger  logging.Logger
}

// NewBalanceRepository returns a new userRepository instance
func NewBalanceRepository(s *session.Session, l logging.Logger) BalanceRepository {
	return &balanceRepository{s, l}
}

func (b *balanceRepository) Create(balance *data.Balance) error {
	dialect := goqu.Dialect(config.Driver)
	ds := dialect.Insert(config.BalanceTable).Rows(
		goqu.Record{config.UserId: balance.UserID, config.IntegerPart: balance.IntegerPart,
			config.FractionPart: balance.FractionPart, config.Currency: balance.Currency})
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		b.logger.Error("An error occurred while generating the SQL: ", err)
		return err
	}
	_, err = b.session.Exec(sqlStr)
	if err != nil {
		b.logger.Error(err)
		return err
	}
	return nil
}
