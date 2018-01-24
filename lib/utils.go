package lib

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// NewUUID ...
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// CreateEndpointKey ...
func CreateEndpointKey(method string, endpoint string) string {
	return strings.ToLower(method + endpoint)
}

// Broadcast ...
func Broadcast(c *gin.Context) {
	url := c.Request.URL
	method := c.Request.Method
	path := url.Path

	var requestBody interface{}
	requestBody = GetRequestBody(c)
	requestHeaders := GetHeaders(c)

	msg := WsMessage{
		Host:   c.Request.Host,
		Body:   requestBody,
		Method: method,
		Path:   path,
		Query:  c.Request.URL.Query(),
		Header: requestHeaders}

	for id, conn := range Clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			conn.Close()
			delete(Clients, id)
		}
	}
}

// GetHeaders ...
func GetHeaders(c *gin.Context) map[string]string {
	hdr := make(map[string]string, len(c.Request.Header))
	for k, v := range c.Request.Header {
		hdr[k] = v[0]
	}
	return hdr
}

// GetIP ...
func GetIP(c *gin.Context) string {
	ip := c.ClientIP()
	return ip
}

// GetMultiPartFormValue ...
func GetMultiPartFormValue(c *gin.Context) interface{} {
	var requestBody interface{}

	multipartForm := make(map[string]interface{})
	if err := c.Request.ParseMultipartForm(DefaultMemory); err != nil {
		// handle error
	}
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			multipartForm[key] = strings.Join(values, "")
		}

		for key, file := range c.Request.MultipartForm.File {
			for k, f := range file {
				formKey := fmt.Sprintf("%s%d", key, k)
				multipartForm[formKey] = map[string]interface{}{"filename": f.Filename, "size": f.Size}
			}
		}

		if len(multipartForm) > 0 {
			requestBody = multipartForm
		}
	}
	return requestBody
}

// GetFormBody ...
func GetFormBody(c *gin.Context) interface{} {
	var requestBody interface{}

	form := make(map[string]string)
	if err := c.Request.ParseForm(); err != nil {
		// handle error
	}
	for key, values := range c.Request.PostForm {
		form[key] = strings.Join(values, "")
	}
	if len(form) > 0 {
		requestBody = form
	}

	return requestBody
}

// TryBind ...
func TryBind(c *gin.Context) interface{} {
	var model interface{}
	err := c.Bind(&model)
	if err != nil {
		return nil
	}
	return model
}

// GetRequestBody ...
func GetRequestBody(c *gin.Context) interface{} {
	multiPartFormValue := GetMultiPartFormValue(c)
	if multiPartFormValue != nil {
		return multiPartFormValue
	}

	formBody := GetFormBody(c)
	if formBody != nil {
		return formBody
	}

	contentType := c.ContentType()
	method := c.Request.Method
	if method == "GET" {
		return nil
	}

	switch contentType {
	case binding.MIMEJSON:
		return TryBind(c)
	case binding.MIMEXML, binding.MIMEXML2:
		var model interface{}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err == nil {
			model = string(body)
		}
		return model
	default:
		return nil
	}
}
