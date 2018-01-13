package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

const httpLiveFileProviderEnvKey = "HttpLiveFileProvider"

// CORSMiddleware ...
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// StaticFileMiddleware ...
func StaticFileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL
		uriPath := url.Path
		method := c.Request.Method
		assetPath := "public" + uriPath
		ext := path.Ext(assetPath)
		if method == "GET" && uriPath == "/" {
			assetPath = "public/index.html"
		}

		if ext == ".map" {
			c.Status(404)
			c.Abort()
			return
		}

		fp := os.Getenv(httpLiveFileProviderEnvKey)
		if fp != "" {
			TryGetLocalFile(c, assetPath)
		} else {
			TryGetAssetFile(c, assetPath)
		}

		if c.IsAborted() {
			return
		}

		c.Next()
	}
}

// APIMiddleware ...
func APIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL
		method := c.Request.Method
		path := url.Path
		key := CreateEndpointKey(method, path)
		model, err := GetEndpoint(key)
		if err == nil && model != nil {

			if model.MimeType != "" {
				c.Data(200, model.MimeType, model.FileContent)
				c.Abort()
				return
			}

			var requestBody interface{}
			requestBody = GetRequestBody(c)
			requestHeaders := GetHeaders(c)

			w := WsMessage{
				Host:   c.Request.Host,
				Body:   requestBody,
				URL:    url.String(), //[scheme:][//[userinfo@]host][/]path[?query][#fragment]
				Method: method,
				Path:   path,
				Header: requestHeaders}
			Broadcast <- w
			go HandleMessages()

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
			fileContent := "define('config', { defaultPort:'" + Environments.DefaultPort + "', savePath: '/webcli/api/save', " +
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
