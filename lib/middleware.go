package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

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
		if err != nil {
			Broadcast(c)
			c.JSON(404, err)
			c.Abort()
			return
		}

		if err == nil && model != nil {
			if model.MimeType != "" {
				reader := bytes.NewReader(model.FileContent)
				http.ServeContent(c.Writer, c.Request, model.Filename, time.Now(), reader)
				c.Abort()
				return
			}

			Broadcast(c)

			var body interface{}
			json.Unmarshal([]byte(model.Body), &body)
			c.JSON(200, body)
			c.Abort()
			return
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
