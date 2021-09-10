package session

import (
	"database/sql"
	"my-bank-service/internal/config"
	"my-bank-service/pkg/logging"
)

const beginStatus = 1

type SessionFactory struct {
	*sql.DB
}

// Session Сессия сессия
type Session struct {
	db           *sql.DB // Own database
	tx           *sql.Tx // Own transaction
	commitSign   int8    // Commit the mark, indicate if the transaction should be committed
	rollbackSign bool    // Rollback flag that determines whether to rollback the transaction
}

// NewSessionFactory creates a session factory
func NewSessionFactory(driverName string) (*SessionFactory, error) {
	logger := logging.GetLogger()
	dataSource := config.LoadConfig(logger)
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}
	factory := new(SessionFactory)
	factory.DB = db
	return factory, nil
}

// GetSession get a session
func (sf *SessionFactory) GetSession() *Session {
	session := new(Session)
	session.db = sf.DB
	return session
}

// Begin starts a transaction
func (s *Session) Begin() error {
	s.rollbackSign = true
	if s.tx == nil {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		s.tx = tx
		s.commitSign = beginStatus
		return nil
	}
	s.commitSign++
	return nil
}

// Rollback rolls back the transaction
func (s *Session) Rollback() error {
	if s.tx != nil && s.rollbackSign == true {
		err := s.tx.Rollback()
		if err != nil {
			return err
		}
		s.tx = nil
		return nil
	}
	return nil
}

// Commit commits a transaction
func (s *Session) Commit() error {
	s.rollbackSign = false
	if s.tx != nil {
		if s.commitSign == beginStatus {
			err := s.tx.Commit()
			if err != nil {
				return err
			}
			s.tx = nil
			return nil
		} else {
			s.commitSign--
		}
		return nil
	}
	return nil
}

// Exec executes a sql statement, if a transaction was opened, it will be executed in transaction mode,
// if a transaction is not open, it will be executed in non-transactional mode
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.Exec(query, args...)
	}
	return s.db.Exec(query, args...)
}

// QueryRow if the transaction was opened, it will be executed in a transactional way, if the transaction
// is not open, it will be executed in a non-transactional way
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRow(query, args...)
	}
	return s.db.QueryRow(query, args...)
}

// Query request for request data, if the transaction was opened, it will be executed in transaction mode,
// if the transaction is not open, it will be executed in non-transactional mode
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.Query(query, args...)
	}
	return s.db.Query(query, args...)
}

// Prepare preliminary execution, if the transaction was opened, it will be executed in a
// transactional way, if the transaction is not open, it will be executed in a non-transactional way
func (s *Session) Prepare(query string) (*sql.Stmt, error) {
	if s.tx != nil {
		return s.tx.Prepare(query)
	}
	return s.db.Prepare(query)
}
