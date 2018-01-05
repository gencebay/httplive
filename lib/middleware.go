package lib

import (
	"encoding/json"
	"fmt"

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
		url := c.Request.URL
		method := c.Request.Method
		path := url.Path

		OpenDb()
		key := CreateEndpointKey(method, path)
		model, err := GetEndpoint(key)
		CloseDb()
		if err == nil && model != nil {
			var body interface{}
			err := json.Unmarshal([]byte(model.Body), &body)
			if err == nil {
				c.JSON(200, body)
				c.Abort()
			} else {
				c.JSON(200, body)
				c.Abort()
			}
		}
		c.Next()
	}
}

// ConfigJsMiddleware ...
func ConfigJsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL
		path := url.Path
		if path == "/config.js" {
			fileContent := "define('config', { port:'" + Environments.Port + "', savePath: '/webcli/api/save', " +
				"fetchPath: '/webcli/api/endpoint', deletePath: '/webcli/api/deleteendpoint', " +
				"treePath: '/webcli/api/tree', componentId: ''});"
			c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileContent)))
			c.Writer.Header().Set("Content-Type", "application/javascript")
			c.String(200, fileContent)
			return
		}
		c.Next()
	}
}
