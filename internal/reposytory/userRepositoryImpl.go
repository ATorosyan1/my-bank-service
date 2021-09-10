package reposytory

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
	uuid "github.com/satori/go.uuid"
	"github.com/uniplaces/carbon"
	"my-bank-service/internal/config"
	"my-bank-service/internal/data"
	"my-bank-service/pkg/logging"
	"my-bank-service/pkg/session"
)

// userRepository has the implementation of the db methods.
type userRepository struct {
	session *session.Session
	logger  logging.Logger
}

// NewUserRepository returns a new userRepository instance
func NewUserRepository(s *session.Session, l logging.Logger) UserRepository {
	return &userRepository{s, l}
}

func (r *userRepository) Create(user *data.User) error {
	user.ID = uuid.NewV4().String()
	user.CreatedAt = carbon.Now().String()
	user.UpdatedAt = carbon.Now().String()

	dialect := goqu.Dialect(config.Driver)
	ds := dialect.Insert(config.UsersTable).Rows(
		goqu.Record{config.Id: user.ID, config.Email: user.Email, config.UserName: user.Username,
			config.Password: user.Password, config.TokenHash: user.TokenHash,
			config.CreatedAt: user.CreatedAt, config.UpdatedAt: user.UpdatedAt})
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		r.logger.Error("An error occurred while generating the SQL: ", err)
		return err
	}
	_, err = r.session.Exec(sqlStr)
	if err != nil {
		r.logger.Error(err)
		return err
	}

	balance := data.Balance{
		UserID:       user.ID,
		IntegerPart:  8,
		FractionPart: 0,
		Currency:     "USD",
	}
	bDs := dialect.Insert(config.BalanceTable).Rows(
		goqu.Record{config.UserId: balance.UserID, config.IntegerPart: balance.IntegerPart,
			config.FractionPart: balance.FractionPart, config.Currency: balance.Currency})
	sqlStr, _, err = bDs.ToSQL()
	if err != nil {
		r.logger.Error("An error occurred while generating the SQL: ", err)
		return err
	}
	_, err = r.session.Exec(sqlStr)
	if err != nil {
		r.logger.Error(err)
		return err
	}
	return nil
}

func (r *userRepository) GetUserByEmail(email string) (*data.User, error) {
	dialect := goqu.Dialect(config.Driver)
	ds := dialect.From(config.UsersTable).Where(goqu.Ex{config.Email: email})
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		r.logger.Error("An error occurred while generating the SQL: ", err)
		return nil, err
	}
	res, err := r.session.Query(sqlStr)
	defer func() {
		err = res.Close()
		if err != nil {
			r.logger.Error(err)
		}
	}()
	if err != nil {
		return nil, err
	}
	var user data.User
	if res.Next() {
		err = res.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.TokenHash,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.logger.Error(err)
			return nil, err
		}
	}
	return &user, nil
}

func (r *userRepository) GetUserByUserName(userName string) (*data.User, error) {
	dialect := goqu.Dialect(config.Driver)
	ds := dialect.From(config.UsersTable).Where(goqu.Ex{config.UserName: userName})
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		r.logger.Error("An error occurred while generating the SQL: ", err)
		return nil, err
	}
	res, err := r.session.Query(sqlStr)
	defer func() {
		err = res.Close()
		if err != nil {
			r.logger.Error(err)
		}
	}()
	if err != nil {
		return nil, err
	}
	var user data.User
	if res.Next() {
		err = res.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.TokenHash,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.logger.Error(err)
			return nil, err
		}
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(userID string) (*data.User, error) {
	dialect := goqu.Dialect(config.Driver)
	ds := dialect.From(config.UsersTable).Where(goqu.Ex{config.Id: userID})
	sqlStr, _, err := ds.ToSQL()
	if err != nil {
		r.logger.Error("An error occurred while generating the SQL: ", err)
		return nil, err
	}
	res, err := r.session.Query(sqlStr)
	defer func() {
		err = res.Close()
		if err != nil {
			r.logger.Error(err)
		}
	}()
	if err != nil {
		return nil, err
	}
	var user data.User
	if res.Next() {
		err = res.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.TokenHash,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
