package redisdb

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func init_viper() {
	viper.SetConfigFile(`configurations/redis.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

}
func NewReddisConn() *redis.Client {
	init_viper()
	var address string
	if viper.GetBool(`docker`){
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
