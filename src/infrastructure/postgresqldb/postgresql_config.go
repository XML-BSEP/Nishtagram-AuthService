package postgresqldb

import (
	"fmt"
	logger "github.com/jelena-vlajkov/logger/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func NewDBConnection(logger *logger.Logger) *gorm.DB {
	return getConnection(logger)
}

func init_viper(logger *logger.Logger) {
	viper.SetConfigFile(`src/configurations/postgresql.json`)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Logger.Infof("error while reading postgresql config file, error: %v\n", err)
	}

	if viper.GetBool(`debug`) {
		logger.Logger.Infof("running in DEBUG mode")
	}
}

func getConnection(logger *logger.Logger) *gorm.DB {
	init_viper(logger)
	var host string
	if os.Getenv("DOCKER_ENV") != "" {
		host = viper.GetString(`database.host_docker`)
	}else{
		host = viper.GetString(`database.host_localhost`)
	}
	port := viper.GetString(`database.port`)
	user := viper.GetString(`database.user`)
	password := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.dbname`)
	connectionString := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		logger.Logger.Fatalf("error while connection on postgreSQL, error: %v\n", err)
	}

	db.Set("gorm:table_options", "ENGINE=InnoDB")
	return db
}
