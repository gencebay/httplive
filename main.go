package main

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

// IPResponse ...
type IPResponse struct {
	Origin string `json:"origin"`
}

// UserAgentResponse ...
type UserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

// HeadersResponse ...
type HeadersResponse struct {
	Headers map[string]string `json:"headers"`
}

// CookiesResponse ...
type CookiesResponse struct {
	Cookies map[string]string `json:"cookies"`
}

// JSONResponse ...
type JSONResponse interface{}

// GetResponse ...
type GetResponse struct {
	Args map[string][]string `json:"args"`
	HeadersResponse
	IPResponse
	URL string `json:"url"`
}

// PostResponse ...
type PostResponse struct {
	Args map[string][]string `json:"args"`
	Data JSONResponse        `json:"data"`
	Form map[string]string   `json:"form"`
	HeadersResponse
	IPResponse
	URL string `json:"url"`
}

// GzipResponse ...
type GzipResponse struct {
	HeadersResponse
	IPResponse
	Gzipped bool `json:"gzipped"`
}

// DeflateResponse ...
type DeflateResponse struct {
	HeadersResponse
	IPResponse
	Deflated bool `json:"deflated"`
}

// BasicAuthResponse ...
type BasicAuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	User          string `json:"string"`
}

// WebCliController ...
type WebCliController struct {
	Port string
}

// APIDataModel ...
type APIDataModel struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Body     string `json:"body"`
}

// JsTreeDataModel ...
type JsTreeDataModel struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Type     string            `json:"type"`
	Children []JsTreeDataModel `json:"children"`
}

var (
	database      = "httpbin.db"
	hostingPort   = "5003"
	httpMethodMap = map[string]string{
		"GET":    "label label-primary label-small",
		"POST":   "label label-success label-small",
		"PUT":    "label label-warning label-small",
		"DELETE": "label label-danger label-small",
	}
)

var db *bolt.DB
var dbOpen bool

// OpenDb ...
func OpenDb() error {
	var err error
	_, filename, _, _ := runtime.Caller(0) // get full path of this file
	dbfile := path.Join(path.Dir(filename), database)
	config := &bolt.Options{Timeout: 1 * time.Second}
	db, err = bolt.Open(dbfile, 0600, config)
	if err != nil {
		log.Fatal(err)
	}
	dbOpen = true
	return nil
}

// CloseDb ...
func CloseDb() {
	dbOpen = false
	db.Close()
}

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
		model, err := getEndpoint(method, path)
		CloseDb()
		if err == nil {
			var body interface{}
			err := json.Unmarshal([]byte(model.Body), &body)
			if err == nil {
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
			fileContent := "define('config', { port:'" + hostingPort + "', savePath: '/webcli/api/save', " +
				"fetchPath: '/webcli/api/endpoint', treePath: '/webcli/api/tree', componentId: ''});"
			c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileContent)))
			c.Writer.Header().Set("Content-Type", "application/javascript")
			c.String(200, fileContent)
			return
		}
		c.Next()
	}
}

func createDbBucket(port string) error {
	if !dbOpen {
		return fmt.Errorf("open db connection first")
	}
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(port))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return err
}

func saveEndpoint(model *APIDataModel) error {
	if !dbOpen {
		return fmt.Errorf("open db connection first")
	}

	if model.Endpoint == "" || model.Method == "" {
		return fmt.Errorf("model endpoint and method could not be empty")
	}

	key := model.Method + model.Endpoint
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(hostingPort))
		enc, err := model.encode()
		if err != nil {
			return fmt.Errorf("could not encode Person %s: %s", model.Endpoint, err)
		}
		err = bucket.Put([]byte(key), enc)
		return err
	})
	return err
}

func getEndpoint(method string, endpoint string) (*APIDataModel, error) {
	if !dbOpen {
		return nil, fmt.Errorf("db must be opened before saving")
	}

	if method == "" || endpoint == "" {
		return nil, fmt.Errorf("model endpoint and method could not be empty")
	}

	key := method + endpoint
	var model *APIDataModel
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(hostingPort))
		k := []byte(key)
		model, err = decode(b.Get(k))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not get APIDataModel with key: %s", key)
		return nil, err
	}
	return model, nil
}

func main() {
	var port string
	app := cli.NewApp()
	app.Name = "httpbin"
	app.Usage = "HTTP Request & Response Service, Mock HTTP"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port, p",
			Value:       "5003",
			Destination: &port,
		},
	}

	app.Action = func(c *cli.Context) error {
		host(port)
		return nil
	}
	app.Run(os.Args)
}

func (model *APIDataModel) encode() ([]byte, error) {
	enc, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decode(data []byte) (*APIDataModel, error) {
	var model *APIDataModel
	err := json.Unmarshal(data, &model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (model *APIDataModel) gobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(model)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte) (*APIDataModel, error) {
	var model *APIDataModel
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func host(port string) {
	hostingPort = port

	OpenDb()
	createDbBucket(port)
	CloseDb()

	r := gin.Default()

	r.Use(CORSMiddleware())

	r.Use(ConfigJsMiddleware())

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	r.Use(APIMiddleware())

	r.POST("/", apiPostHandler)
	r.GET("/ip", ipHandler)
	r.GET("/uuid", uuidHandler)
	r.GET("/user-agent", userAgentHandler)
	r.GET("/headers", headersHandler)
	r.GET("/get", getHandler)
	r.POST("/post", postHandler)

	webcli := r.Group("/webcli")
	{
		ctrl := new(WebCliController)
		webcli.GET("/api/tree", ctrl.tree)
		webcli.GET("/api/endpoint", ctrl.endpoint)
		webcli.POST("/api/save", ctrl.save)
	}

	r.NoRoute(func(c *gin.Context) {
		method := c.Request.Method

		if method == "POST" {
			genericPostHandler(c, true)
			return
		}

		c.Status(404)
		c.File("./public/404.html")
	})

	r.Run(":" + port)
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
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
	response := UserAgentResponse{UserAgent: agent}
	c.JSON(200, response)
}

func headersHandler(c *gin.Context) {
	response := HeadersResponse{Headers: getHeaders(c)}
	c.JSON(200, response)
}

func getHandler(c *gin.Context) {
	ip := c.ClientIP()
	url := c.Request.RequestURI
	headers := getHeaders(c)
	args := c.Request.URL.Query()
	response := GetResponse{
		Args:            args,
		HeadersResponse: HeadersResponse{Headers: headers},
		URL:             url,
		IPResponse:      IPResponse{Origin: ip},
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

	if json != nil {
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
	}

	response := PostResponse{
		Args:            args,
		HeadersResponse: HeadersResponse{Headers: headers},
		URL:             url,
		IPResponse:      IPResponse{Origin: ip},
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

func createJsTreeModel(a APIDataModel) JsTreeDataModel {

	model := JsTreeDataModel{ID: a.Endpoint, Text: a.Endpoint, Children: []JsTreeDataModel{}}
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

func (ctrl WebCliController) tree(c *gin.Context) {
	apis := []APIDataModel{
		{Endpoint: "/api/token/mobiletoken", Method: "GET"},
		{Endpoint: "/api/users/list", Method: "GET"},
		{Endpoint: "/api/users/create", Method: "POST"},
		{Endpoint: "/api/users/update", Method: "PUT"},
		{Endpoint: "/api/users/delete", Method: "DELETE"},
	}
	trees := []JsTreeDataModel{}
	for _, api := range apis {
		trees = append(trees, createJsTreeModel(api))
	}

	state := map[string]interface{}{
		"opened": true,
	}

	c.JSON(200, gin.H{
		"id":       "APIs",
		"text":     "APIs",
		"state":    state,
		"children": trees,
		"type":     "root",
	})
}

func (ctrl WebCliController) endpoint(c *gin.Context) {
	query := c.Request.URL.Query()
	endpoint := query.Get("endpoint")
	method := query.Get("method")
	if endpoint == "" || method == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint and method required"})
		return
	}

	OpenDb()
	model, err := getEndpoint(method, endpoint)
	CloseDb()
	if err != nil {
		c.JSON(http.StatusOK, model)
		return
	}

	c.JSON(200, model)
}

func (ctrl WebCliController) save(c *gin.Context) {
	var model APIDataModel
	if err := c.ShouldBindJSON(&model); err == nil {
		OpenDb()
		err := saveEndpoint(&model)
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

func (ctrl WebCliController) saveapi(c *gin.Context) {
	var model APIDataModel
	if err := c.ShouldBindJSON(&model); err == nil {
		OpenDb()
		err := saveEndpoint(&model)
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
