package config

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tkanos/gonfig"
	"my-bank-service/pkg/logging"
	"sync"
)

const (
	AccessTokenLife = 30
)

const (
	ServerAddr  = ""
	ServerPort  = "8080"
	Driver      = "mysql"
	DatabaseUrl = "%s:%s@/%s?parseTime=true"

	LogConfigFileName = "logConfig"
	ServerConfigPath  = "./properties"
	DbConfigPath      = "./properties/dbConfig.yml"
)

const (
	GroupPath        = "/"
	SignUpPath       = "user/add/"
	LoginPath        = "login/"
	LogoutPath       = "logout/"
	RefreshTokenPath = "/refresh-token/"
)

const (
	UsersTable   = "users"
	BalanceTable = "balance"

	Id        = "id"
	Email     = "email"
	UserName  = "userName"
	Password  = "password"
	TokenHash = "tokenhash"
	CreatedAt = "createdat"
	UpdatedAt = "updatedat"

	UserId       = "user_id"
	IntegerPart  = "integer_part"
	FractionPart = "fraction_part"
	Currency     = "currency"
)

var instance *logging.Configuration
var logOnce sync.Once

// Configurations wraps all the configs variables required by the auth service
type Configurations struct {
	AccessTokenPrivateKeyPath  string
	AccessTokenPublicKeyPath   string
	RefreshTokenPrivateKeyPath string
	RefreshTokenPublicKeyPath  string
	JwtExpiration              int // in minutes
	SendGridApiKey             string
}

// NewConfigurations returns a new Configuration object
func NewConfigurations(logger logging.Logger) *Configurations {

	viper.AutomaticEnv()

	logger.Debug("found database url in env, connection string is formed by parsing it")

	viper.SetDefault("ACCESS_TOKEN_PRIVATE_KEY_PATH", "./internal/access-private.pem") //"./web-service/d-link_snmp/snmpAPI/pkg/jwt/access-private.pem")
	viper.SetDefault("ACCESS_TOKEN_PUBLIC_KEY_PATH", "./internal/access-public.pem")
	viper.SetDefault("REFRESH_TOKEN_PRIVATE_KEY_PATH", "./internal/refresh-private.pem")
	viper.SetDefault("REFRESH_TOKEN_PUBLIC_KEY_PATH", "./internal/refresh-public.pem")
	viper.SetDefault("JWT_EXPIRATION", AccessTokenLife)

	configs := &Configurations{
		JwtExpiration:              viper.GetInt("JWT_EXPIRATION"),
		AccessTokenPrivateKeyPath:  viper.GetString("ACCESS_TOKEN_PRIVATE_KEY_PATH"),
		AccessTokenPublicKeyPath:   viper.GetString("ACCESS_TOKEN_PUBLIC_KEY_PATH"),
		RefreshTokenPrivateKeyPath: viper.GetString("REFRESH_TOKEN_PRIVATE_KEY_PATH"),
		RefreshTokenPublicKeyPath:  viper.GetString("REFRESH_TOKEN_PUBLIC_KEY_PATH"),
		SendGridApiKey:             viper.GetString("SENDGRID_API_KEY"),
	}

	return configs
}

// GetLogConfiguration reads log configuration from a config file
func GetLogConfiguration() *logging.Configuration {
	logOnce.Do(func() {
		viper.SetConfigName(LogConfigFileName)
		viper.AddConfigPath(ServerConfigPath)
		err := viper.ReadInConfig() // Find and read the uploaderConfig file
		if err != nil {             // Handle errors reading the uploaderConfig file
			logrus.Error(err)
		}
		err = viper.Unmarshal(&instance)
		if err != nil {
			logrus.Error(err)
		}
	})
	return instance
}

var connection string
var dbOnce sync.Once

// LoadConfig get sql connection parameters
func LoadConfig(l logging.Logger) string {
	dbOnce.Do(func() {
		config := mysql.Config{}
		err := gonfig.GetConf(DbConfigPath, &config)
		if err != nil {
			l.Logger.Error("An error was generated while reading the database config file.")
			return
		}
		connection = fmt.Sprintf(DatabaseUrl, config.User, config.Passwd, config.DBName)
	})
	return connection
}
