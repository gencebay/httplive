package main

import (
	"fmt"
	"net/http"

	"github.com/gencebay/httpbin/controllers"
	"github.com/gencebay/httpbin/types"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware ...
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func main() {
	r := gin.Default()

	r.Use(CORSMiddleware())

	v1 := r.Group("/api")
	{
		echoCtrl := new(controllers.EchoController)

		v1.GET("/echo", echoCtrl.Echo)
		v1.POST("/echo", echoCtrl.Echo)
	}

	r.LoadHTMLGlob("./wwwroot/*")

	r.Static("/wwwroot", "./wwwroot")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"version": "0.0.1",
			// "goVersion":             runtime.Version(),
		})
	})

	r.GET("/ip", ipHandler)
	r.GET("/user-agent", userAgentHandler)
	r.GET("/headers", headersHandler)
	r.GET("/get", getHandler)

	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{})
	})

	r.Run(":8080")
}

func getHeaders(c *gin.Context) map[string]string {
	hdr := make(map[string]string, len(c.Request.Header))
	for k, v := range c.Request.Header {
		hdr[k] = v[0]
	}
	return hdr
}

func ipHandler(c *gin.Context) {
	ip := c.ClientIP()
	c.JSON(200, gin.H{
		"origin": ip,
	})
}

func userAgentHandler(c *gin.Context) {
	agent := c.GetHeader("User-Agent")
	response := types.UserAgentResponse{UserAgent: agent}
	c.JSON(200, response)
}

func headersHandler(c *gin.Context) {
	response := types.HeadersResponse{Headers: getHeaders(c)}
	c.JSON(200, response)
}

func getHandler(c *gin.Context) {
	ip := c.ClientIP()
	url := c.Request.RequestURI
	headers := getHeaders(c)
	args := c.Request.URL.Query()
	response := types.GetResponse{
		Args:            args,
		HeadersResponse: types.HeadersResponse{Headers: headers},
		URL:             url,
		IPResponse:      types.IPResponse{Origin: ip},
	}
	c.JSON(200, response)
}
