module auth-service

go 1.16

replace github.com/jelena-vlajkov/logger/logger => ../../Nishtagram-Logger/

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/casbin/casbin/v2 v2.31.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/secure v0.0.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-redis/redis/v8 v8.8.3
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.2.0
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jelena-vlajkov/logger/logger v1.0.0
	github.com/microcosm-cc/bluemonday v1.0.10
	github.com/myesui/uuid v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/pquerna/otp v1.3.0
	github.com/spf13/viper v1.7.1
	github.com/twinj/uuid v1.0.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	google.golang.org/genproto v0.0.0-20210614182748-5b3b54cad159 // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/stretchr/testify.v1 v1.2.2 // indirect
	gorm.io/driver/postgres v1.1.0
	gorm.io/gorm v1.21.10

)
