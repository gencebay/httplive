package lib

import (
	"log"
	"mime"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

// TryGetLocalFile ...
func TryGetLocalFile(c *gin.Context, filePath string) {
	log.Printf("fs:dev local file for: %s", filePath)
	f := path.Join(Environments.WorkingDirectory, filePath)
	if _, err := os.Stat(f); err == nil {
		c.File(f)
		c.Abort()
		return
	}
}

// TryGetAssetFile ...
func TryGetAssetFile(c *gin.Context, filePath string) {
	log.Printf("fs:bindata asset trygetfile executed for: %s", filePath)
	assetData, err := Asset(filePath)
	if err == nil && assetData != nil {
		ext := path.Ext(filePath)
		contentType := mime.TypeByExtension(ext)
		c.Data(200, contentType, assetData)
		c.Abort()
		return
	}
}
