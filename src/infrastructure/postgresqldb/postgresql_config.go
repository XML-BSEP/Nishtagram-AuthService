package postgresqldb

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func NewDBConnection() *gorm.DB {
	return getConnection()
}

func init_viper() {
	viper.SetConfigFile(`configurations/postgresql.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func getConnection() *gorm.DB {
	init_viper()
	var host string
	if viper.GetBool(`docker`){
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
		panic(err)
	}

	db.Set("gorm:table_options", "ENGINE=InnoDB")
	return db
}
