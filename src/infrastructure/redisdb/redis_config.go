package redisdb

import (
	"github.com/go-redis/redis/v8"
	logger "github.com/jelena-vlajkov/logger/logger"
	"github.com/spf13/viper"
	"os"
)

func init_viper(logger *logger.Logger) {
	viper.SetConfigFile(`src/configurations/redis.json`)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Logger.Fatalf("errow while reading redis config file, error: %v\n", err)
	}

}
func NewReddisConn(logger *logger.Logger) *redis.Client {
	init_viper(logger)
	var address string
	if os.Getenv("DOCKER_ENV") != "" {
		address = viper.GetString(`server.address_docker`)
	}else{
		address = viper.GetString(`server.address_localhost`)
	}
	port := viper.GetString(`server.port`)

	return redis.NewClient(&redis.Options{
		Addr: address + ":" + port,
		Password: "",
		DB: 0,
	})
}
