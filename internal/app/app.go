package app

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"my-bank-service/internal/config"
	"my-bank-service/internal/data"
	"my-bank-service/internal/handler"
	"my-bank-service/internal/service"
	"my-bank-service/pkg/logging"
	"my-bank-service/pkg/session"
	"os"
)

var sf *session.SessionFactory

func Run(address string, port string) {

	logging.Init()
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	accountHandler := handler.NewAccountHandler(logger, sf)
	accountHandler.Routes(router)

	err := router.Run(fmt.Sprintf("%s:%v", address, port))
	if err != nil {
		logger.Error(err)
	}

}

func init() {
	err := ifNotGenerateTheDatabase()
	if err != nil {
		log.Panic(err)
	}
	sf, err = session.NewSessionFactory(config.Driver, config.DbName)
	if err != nil {
		log.Panic(err)
	}
}

func ifNotGenerateTheDatabase() error {
	fileName := fmt.Sprintf(config.DbName)
	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		log.Println("Creating myBank.db...")
		file, err := os.Create(config.DbName)
		if err != nil {
			return err
		}
		file.Close()
		log.Println("myBank.db created")

		sqliteDatabase, _ := sql.Open(config.Driver, fmt.Sprintf("./%s", config.DbName))
		defer sqliteDatabase.Close()
		err = createTables(sqliteDatabase)
		if err != nil {
			return err
		}
		account := data.Account{}
		err = inserts(sqliteDatabase, account)
		if err != nil {
			return err
		}
	}
	return nil
}

func createTables(db *sql.DB) error {
	createAccountTableSQL := `CREATE TABLE account (
		"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"balanceId" INTEGER NOT NULL,
		"currency" TEXT NOT NULL
	  );` // SQL Statement for Create Table
	createBalanceTableSQL := `CREATE TABLE balance (
		"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"sbp" NUMERIC NOT NULL,
		"spf" NUMERIC NOT NULL
	  );`

	log.Println("Create account table...")
	statement, err := db.Prepare(createAccountTableSQL)
	if err != nil {
		return err
	}
	statement.Exec() // Execute SQL Statements
	log.Println("account table created")

	log.Println("Create balance table...")
	statement, err = db.Prepare(createBalanceTableSQL)
	if err != nil {
		return err
	}
	statement.Exec() // Execute SQL Statements
	log.Println("balance table created")
	return nil
}

func inserts(db *sql.DB, account data.Account) error {
	log.Println("Inserting balance record ...")
	insertBalanceSQL := `INSERT INTO balance(sbp, spf) VALUES (?, ?)`
	statement, err := db.Prepare(insertBalanceSQL)

	if err != nil {
		return err
	}
	_, err = statement.Exec(account.Balance.SBP, account.Balance.SPF)
	if err != nil {
		return err
	}

	log.Println("Inserting balance record ...")
	insertAccountSQL := `INSERT INTO account(balanceId, currency) VALUES (?, ?)`
	statement, err = db.Prepare(insertAccountSQL)

	if err != nil {
		return err
	}
	_, err = statement.Exec(1, service.SBP)
	if err != nil {
		return err
	}
	return nil
}
