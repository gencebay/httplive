package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	. "github.com/gencebay/httplive/lib"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

func main() {
	var ports string
	var dbpath string
	app := cli.NewApp()
	app.Name = "httplive"
	app.Usage = "HTTP Request & Response Service, Mock HTTP"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "ports, p",
			Value:       "5003",
			Usage:       "Hosting ports can be array with semicolon <5003,5004> to host multiple endpoint. To array usage <dbpath> flag required. First one is DefaultPort (DbKey)",
			Destination: &ports,
		},
		cli.StringFlag{
			Name:        "dbpath, d",
			Value:       "",
			Usage:       "Fullpath of the httplive.db with forward slash.",
			Destination: &dbpath,
		},
	}

	app.Action = func(c *cli.Context) error {
		host(ports, dbpath)
		return nil
	}
	app.Run(os.Args)
}

func createDb() error {
	var err error
	_, filename, _, _ := runtime.Caller(0) // get full path of this file
	var dbfile string
	if Environments.DatabaseAttachedFullPath != "" {
		if _, err := os.Stat(Environments.DatabaseAttachedFullPath); os.IsNotExist(err) {
			log.Fatal(err)
		}
		dbfile = Environments.DatabaseAttachedFullPath
	} else {
		dbfile = path.Join(path.Dir(filename), Environments.DatabaseName)
	}

	Environments.DbFile = dbfile
	return err
}

func host(ports string, dbPath string) {

	portsArr := strings.Split(ports, ",")
	port := portsArr[0]
	length := len(portsArr)
	hasMultiplePort := false
	if length > 1 && dbPath != "" {
		hasMultiplePort = true
	}

	Environments.DefaultPort = port
	Environments.HasMultiplePort = hasMultiplePort
	Environments.DatabaseAttachedFullPath = dbPath

	createDb()

	OpenDb()
	CreateDbBucket()
	CloseDb()

	OpenDb()
	InitDbValues()
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
		webcli.GET("/api/backup", ctrl.Backup)
		webcli.GET("/api/tree", ctrl.Tree)
		webcli.GET("/api/endpoint", ctrl.Endpoint)
		webcli.GET("/api/deleteendpoint", ctrl.DeleteEndpoint)
		webcli.POST("/api/save", ctrl.Save)
		webcli.POST("/api/saveendpoint", ctrl.SaveEndpoint)
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

	if hasMultiplePort {
		for i := 1; i < length; i++ {
			go func(port string) {
				r.Run(":" + port)
			}(portsArr[i])
		}
	}

	r.Run(":" + port)
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
	uuid, err := NewUUID()
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
