package lib

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

var (
	httpMethodMap = map[string]string{
		"GET":    "label label-primary label-small",
		"POST":   "label label-success label-small",
		"PUT":    "label label-warning label-small",
		"DELETE": "label label-danger label-small",
	}
)

func createJsTreeModel(a APIDataModel) JsTreeDataModel {
	model := JsTreeDataModel{ID: a.ID, Key: a.Endpoint, Text: a.Endpoint, Children: []JsTreeDataModel{}}
	endpointText := `<span class="%v">%v</span> %v`
	switch method := a.Method; method {
	case "GET":
		model.Type = "GET"
		model.Text = fmt.Sprintf(endpointText, httpMethodMap["GET"], "GET", a.Endpoint)
	case "POST":
		model.Type = "POST"
		model.Text = fmt.Sprintf(endpointText, httpMethodMap["POST"], "POST", a.Endpoint)
	case "PUT":
		model.Type = "PUT"
		model.Text = fmt.Sprintf(endpointText, httpMethodMap["PUT"], "PUT", a.Endpoint)
	case "DELETE":
		model.Type = "DELETE"
		model.Text = fmt.Sprintf(endpointText, httpMethodMap["DELETE"], "DELETE", a.Endpoint)
	default:
		model.Type = "GET"
		model.Text = fmt.Sprintf(endpointText, httpMethodMap["GET"], "GET", a.Endpoint)
	}
	return model
}

// Tree ...
func (ctrl WebCliController) Tree(c *gin.Context) {
	trees := []JsTreeDataModel{}
	OpenDb()
	apis := EndpointList()
	CloseDb()
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
	err := db.View(func(tx *bolt.Tx) error {
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+Environments.DatabaseName+`"`)
		c.Writer.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(c.Writer)
		CloseDb()
		return err
	})
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
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
	OpenDb()
	model, err := GetEndpoint(key)
	CloseDb()
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
		OpenDb()
		err := SaveEndpoint(&model)
		CloseDb()
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
	if err := c.ShouldBindJSON(&model); err == nil {
		key := model.Key
		if key != "" {
			// try update endpoint
			OpenDb()
			endpoint, _ := GetEndpoint(key)
			CloseDb()
			if endpoint != nil {
				endpoint.Method = model.Method
				endpoint.Endpoint = model.Endpoint
				OpenDb()
				DeleteEndpoint(key)
				err := SaveEndpoint(endpoint)
				CloseDb()
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				}
			}
		} else {
			// new endpoint
			endpoint := APIDataModel{Endpoint: model.Endpoint, Method: model.Method}
			OpenDb()
			err := SaveEndpoint(&endpoint)
			CloseDb()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	OpenDb()
	DeleteEndpoint(key)
	CloseDb()

	c.JSON(200, gin.H{
		"success": "ok",
	})
}
