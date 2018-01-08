package lib

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
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

// HandleMessages ...
func HandleMessages() {
	for {
		msg := <-Broadcast
		// Bağlı tüm kullanıcılara gönder
		for client := range Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(Clients, client)
			}
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

	multipartForm := make(map[string]string)
	if err := c.Request.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		// handle error
	}
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			multipartForm[key] = strings.Join(values, "")
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

// GetBodyJSON ...
func GetBodyJSON(c *gin.Context) interface{} {
	var requestBody interface{}

	var jsonModel interface{}
	c.BindJSON(&jsonModel)
	if jsonModel != nil {
		requestBody = jsonModel
	}
	return requestBody
}

// GetRequestBody ...
func GetRequestBody(c *gin.Context) interface{} {
	var requestBody interface{}

	multiPartFormValue := GetMultiPartFormValue(c)
	if multiPartFormValue != nil {
		requestBody = multiPartFormValue
		return requestBody
	}

	formBody := GetFormBody(c)
	if formBody != nil {
		requestBody = formBody
		return requestBody
	}

	bodyJSON := GetBodyJSON(c)
	if bodyJSON != nil {
		requestBody = bodyJSON
		return requestBody
	}

	return requestBody
}
