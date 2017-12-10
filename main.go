package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// APIMiddleware ...
func APIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			genericPostHandler(c, true)
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
	r.GET("/uuid", uuidHandler)
	r.GET("/user-agent", userAgentHandler)
	r.GET("/headers", headersHandler)
	r.GET("/get", getHandler)
	r.POST("/", apiPostHandler)
	r.POST("/post", postHandler)

	r.NoRoute(func(c *gin.Context) {
		method := c.Request.Method

		if method == "POST" {
			genericPostHandler(c, true)
			return
		}

		c.HTML(404, "404.html", gin.H{"method": method})
	})

	r.Run(":8080")
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
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

func uuidHandler(c *gin.Context) {
	uuid, err := newUUID()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	c.JSON(200, gin.H{
		"uuid": uuid,
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

func genericPostHandler(c *gin.Context, mockReturn bool) {
	ip := c.ClientIP()
	url := c.Request.RequestURI
	headers := getHeaders(c)
	args := c.Request.URL.Query()
	form := make(map[string]string)
	if err := c.Request.ParseForm(); err != nil {
		// handle error
	}

	for key, values := range c.Request.PostForm {
		form[key] = strings.Join(values, "")
	}

	if mockReturn {
		i, ok := form["return"]
		if ok {
			byt := []byte(i)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				c.JSON(400, "Invalid JSON format")
				return
			}

			c.JSON(200, dat)
			return
		}
	}

	var json interface{}
	c.BindJSON(&json)

	v, ok := json.(map[string]interface{})
	if !ok {
		c.JSON(400, "Invalid JSON format")
		return
	}

	obj, ok := v["return"]
	if ok {
		c.JSON(200, obj)
		return
	}

	response := types.PostResponse{
		Args:            args,
		HeadersResponse: types.HeadersResponse{Headers: headers},
		URL:             url,
		IPResponse:      types.IPResponse{Origin: ip},
		Form:            form,
		Data:            json,
	}

	c.JSON(200, response)
}

func apiPostHandler(c *gin.Context) {
	genericPostHandler(c, true)
}

func postHandler(c *gin.Context) {
	genericPostHandler(c, false)
}
