package session

import (
	"database/sql"
	"fmt"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

const beginStatus = 1

// SessionFactory Фабрика сеансов
type SessionFactory struct {
	*sql.DB
}

// Session Сессия сессия
type Session struct {
	db           *sql.DB // Собственная база данных
	tx           *sql.Tx // Собственная транзакция
	commitSign   int8    // Зафиксировать отметку, указать, следует ли фиксировать транзакцию
	rollbackSign bool    // Флаг отката, определяющий, откатывать ли транзакцию
}

// NewSessionFactory создает фабрику сеансов
func NewSessionFactory(driverName, dataSourceName string) (*SessionFactory, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		fmt.Println(err)
	}
	factory := new(SessionFactory)
	factory.DB = db
	return factory, nil
}

// GetSession Получить сеанс
func (sf *SessionFactory) GetSession() *Session {
	session := new(Session)
	session.db = sf.DB
	return session
}

// Begin запускает транзакцию
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

// Rollback Откат откатывает транзакцию
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

// Commit Фиксация фиксирует транзакцию
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

// Exec выполняет оператор sql, если транзакция была открыта, она будет выполнена в режиме транзакции, если транзакция не открыта, она будет выполнена в нетранзакционном режиме
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.Exec(query, args...)
	}
	return s.db.Exec(query, args...)
}

// QueryRow Если транзакция была открыта, она будет выполняться транзакционным способом, если транзакция не открыта, она будет выполнена нетранзакционным способом
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRow(query, args...)
	}
	return s.db.QueryRow(query, args...)
}

// Query Запрос данных запроса, если транзакция была открыта, она будет выполнена в режиме транзакции, если транзакция не открыта, она будет выполнена в нетранзакционном режиме
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.Query(query, args...)
	}
	return s.db.Query(query, args...)
}

// Prepare Подготовить предварительное выполнение, если транзакция была открыта, она будет выполнена транзакционным способом, если транзакция не открыта, она будет выполнена нетранзакционным способом
func (s *Session) Prepare(query string) (*sql.Stmt, error) {
	if s.tx != nil {
		return s.tx.Prepare(query)
	}
	return s.db.Prepare(query)
}
