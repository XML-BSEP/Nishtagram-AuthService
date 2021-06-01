package main

import (
	router2 "auth-service/src/http/router"
	"auth-service/src/infrastructure/postgresqldb"
	"auth-service/src/infrastructure/seeder"
	interactor2 "auth-service/src/interactor"
)

func main() {


	postgreConn := postgresqldb.NewDBConnection()

	seeder.SeedData(postgreConn)

	interactor := interactor2.NewInteractor(postgreConn)
	appHandler := interactor.NewAppHandler()

	router := router2.NewRouter(appHandler)

	/*router := gin.Default()

	router.GET("/generateJWT", func(c *gin.Context) {
		token, err := tuc.CreateToken(12)
		if err != nil {
			c.JSON(400, "Nema")
			c.Abort()
		}
		c.SetCookie("token", token.AccessToken, 10, "/", "127.0.0.1", true, true)
	})

	router.POST("/validateJWT", func(c *gin.Context) {
		//byteBody, _ := ioutil.ReadAll(c.Request.Body)
		s, err := c.Cookie("token")

		if err != nil {
			c.JSON(400, "Ne mere procitat kuki")
			c.Abort()
			return
		}

		if err := tuc.ValidateToken(s); err != nil {
			c.JSON(401, "Token invalid")
			c.Abort()
			return
		}
		c.JSON(200, s)
	})*/



	router.Run("127.0.0.1:8081")
}
