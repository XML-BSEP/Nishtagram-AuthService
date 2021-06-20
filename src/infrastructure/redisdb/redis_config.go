package redisdb

import (
	"github.com/go-redis/redis/v8"
	logger "github.com/jelena-vlajkov/logger/logger"
	"github.com/spf13/viper"
)

func init_viper(logger *logger.Logger) {
	viper.SetConfigFile(`configurations/redis.json`)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Logger.Fatalf("errow while reading redis config file, error: %v\n", err)
	}

}
func NewReddisConn(logger *logger.Logger) *redis.Client {
	init_viper(logger)
	var address string
	var port string
	if viper.GetBool(`docker`){
		address = viper.GetString(`server.address_docker`)
		port = viper.GetString(`server.port_docker`)
	}else{
		address = viper.GetString(`server.address_localhost`)
		port = viper.GetString(`server.port_localhost`)
	}

	return redis.NewClient(&redis.Options{
		Addr: address + ":" + port,
		Password: "",
		DB: 0,
	})
}
