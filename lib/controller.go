package lib

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

var (
	httpMethodLabelMap = map[string]string{
		"GET":    "label label-primary label-small",
		"POST":   "label label-success label-small",
		"PUT":    "label label-warning label-small",
		"DELETE": "label label-danger label-small",
	}
)

func createJsTreeModel(a APIDataModel) JsTreeDataModel {
	originKey := CreateEndpointKey(a.Method, a.Endpoint)
	model := JsTreeDataModel{ID: a.ID, OriginKey: originKey, Key: a.Endpoint, Text: a.Endpoint, Children: []JsTreeDataModel{}}
	endpointText := `<span class="%v">%v</span> %v`
	switch method := a.Method; method {
	case "GET":
		model.Type = "GET"
		model.Text = fmt.Sprintf(endpointText, httpMethodLabelMap["GET"], "GET", a.Endpoint)
	case "POST":
		model.Type = "POST"
		model.Text = fmt.Sprintf(endpointText, httpMethodLabelMap["POST"], "POST", a.Endpoint)
	case "PUT":
		model.Type = "PUT"
		model.Text = fmt.Sprintf(endpointText, httpMethodLabelMap["PUT"], "PUT", a.Endpoint)
	case "DELETE":
		model.Type = "DELETE"
		model.Text = fmt.Sprintf(endpointText, httpMethodLabelMap["DELETE"], "DELETE", a.Endpoint)
	default:
		model.Type = "GET"
		model.Text = fmt.Sprintf(endpointText, httpMethodLabelMap["GET"], "GET", a.Endpoint)
	}
	return model
}

// Tree ...
func (ctrl WebCliController) Tree(c *gin.Context) {
	trees := []JsTreeDataModel{}
	apis := EndpointList()
	for _, api := range apis {
		trees = append(trees, createJsTreeModel(api))
	}

	state := map[string]interface{}{
		"opened": true,
	}

	c.JSON(200, gin.H{
		"id":       "0",
		"key":      "APIs",
		"text":     "APIs",
		"state":    state,
		"children": trees,
		"type":     "root",
	})
}

// Backup ...
func (ctrl WebCliController) Backup(c *gin.Context) {
	OpenDb()
	defer CloseDb()
	err := db.View(func(tx *bolt.Tx) error {
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+Environments.DatabaseName+`"`)
		c.Writer.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(c.Writer)
		return err
	})
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// DownloadFile ...
func (ctrl WebCliController) DownloadFile(c *gin.Context) {
	query := c.Request.URL.Query()
	endpoint := query.Get("endpoint")
	if endpoint != "" {
		key := CreateEndpointKey("GET", endpoint)
		model, err := GetEndpoint(key)
		if err == nil && model != nil {
			if model.MimeType != "" {
				c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+model.Filename+`"`)
				c.Data(200, model.MimeType, model.FileContent)
				return
			}
		}
	}

	c.Status(404)
}

// Endpoint ...
func (ctrl WebCliController) Endpoint(c *gin.Context) {
	query := c.Request.URL.Query()
	endpoint := query.Get("endpoint")
	method := query.Get("method")
	if endpoint == "" || method == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint and method required"})
		return
	}

	key := CreateEndpointKey(method, endpoint)
	model, err := GetEndpoint(key)
	if err != nil {
		c.JSON(http.StatusOK, model)
		return
	}

	c.JSON(200, model)
}

// Save ...
func (ctrl WebCliController) Save(c *gin.Context) {
	var model APIDataModel
	if err := c.ShouldBindJSON(&model); err == nil {
		err := SaveEndpoint(&model)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(200, gin.H{
		"success": "ok",
	})
}

// SaveEndpoint ...
func (ctrl WebCliController) SaveEndpoint(c *gin.Context) {
	var model EndpointModel
	if err := c.ShouldBind(&model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	var mimeType string
	var filename string
	var fileContent []byte
	if model.IsFileResult {
		file, err := c.FormFile("file")
		if err != nil || file == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		f, err := file.Open()
		fileContent, err = ioutil.ReadAll(f)
		mimeType = mime.TypeByExtension(path.Ext(file.Filename))
		filename = file.Filename
	}

	key := model.OriginKey
	if key != "" {
		// try update endpoint
		endpoint, _ := GetEndpoint(key)
		if endpoint != nil {
			endpoint.Method = model.Method
			endpoint.Endpoint = model.Endpoint
			endpoint.MimeType = mimeType
			endpoint.FileContent = fileContent
			endpoint.Filename = filename
			if filename != "" {
				if strings.HasSuffix(endpoint.Endpoint, "/") {
					endpoint.Endpoint += filename
				} else {
					endpoint.Endpoint += "/" + filename
				}
			}
			DeleteEndpoint(key)
			err := SaveEndpoint(endpoint)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}
	} else {
		// new endpoint
		endpoint := APIDataModel{
			Endpoint:    model.Endpoint,
			Method:      model.Method,
			Filename:    filename,
			MimeType:    mimeType,
			FileContent: fileContent}

		if filename != "" {
			if strings.HasSuffix(endpoint.Endpoint, "/") {
				endpoint.Endpoint += filename
			} else {
				endpoint.Endpoint += "/" + filename
			}
		}

		err := SaveEndpoint(&endpoint)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	}

	c.JSON(200, gin.H{
		"success": "ok",
	})
}

// DeleteEndpoint ...
func (ctrl WebCliController) DeleteEndpoint(c *gin.Context) {
	query := c.Request.URL.Query()
	endpoint := query.Get("endpoint")
	method := query.Get("method")
	if endpoint == "" || method == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint and method required"})
		return
	}

	key := CreateEndpointKey(method, endpoint)
	DeleteEndpoint(key)

	c.JSON(200, gin.H{
		"success": "ok",
	})
}
