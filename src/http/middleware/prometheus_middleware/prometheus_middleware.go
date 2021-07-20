package prometheus_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
)

func GetHttpRequestsCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nishtagram_auth_processed_ops_total",
		Help: "The total number of processed events",
	}, []string{"method", "path", "statuscode"})

}


func PrometheusMiddleware(requestsCounter *prometheus.CounterVec)  gin.HandlerFunc {

	return func(c *gin.Context) {

		method := c.Request.Method
		path := c.Request.RequestURI
		statusCode := strconv.Itoa(c.Writer.Status())
		c.Next()

		requestsCounter.With(prometheus.Labels{"method" : method, "path" : path, "statuscode" : statusCode}).Inc()

	}
}

func PrometheusGinHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

