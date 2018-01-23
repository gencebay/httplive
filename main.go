package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	. "github.com/gencebay/httplive/lib"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
			Usage:       "Hosting ports can be array with semicolon <5003,5004> to host multiple endpoint. First one is DefaultPort.",
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
	var dbfile string
	if Environments.DatabaseAttachedFullPath != "" {
		if _, err := os.Stat(Environments.DatabaseAttachedFullPath); os.IsNotExist(err) {
			log.Fatal(err)
		}
		dbfile = Environments.DatabaseAttachedFullPath
	} else {
		dbfile = path.Join(Environments.WorkingDirectory, Environments.DatabaseName)
	}

	Environments.DbFile = dbfile
	CreateDbBucket()
	return err
}

func host(ports string, dbPath string) {

	portsArr := strings.Split(ports, ",")
	port := portsArr[0]
	length := len(portsArr)
	hasMultiplePort := false
	if length > 1 {
		hasMultiplePort = true
	}

	_, filename, _, _ := runtime.Caller(0)
	Environments.WorkingDirectory = path.Dir(filename)
	Environments.DefaultPort = port
	Environments.HasMultiplePort = hasMultiplePort
	Environments.DatabaseAttachedFullPath = dbPath

	createDb()

	InitDbValues()

	r := gin.Default()

	r.Use(StaticFileMiddleware())

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})

	r.Use(CORSMiddleware())

	r.Use(ConfigJsMiddleware())

	webcli := r.Group("/webcli")
	{
		ctrl := new(WebCliController)
		webcli.GET("/api/backup", ctrl.Backup)
		webcli.GET("/api/downloadfile", ctrl.DownloadFile)
		webcli.GET("/api/tree", ctrl.Tree)
		webcli.GET("/api/endpoint", ctrl.Endpoint)
		webcli.GET("/api/deleteendpoint", ctrl.DeleteEndpoint)
		webcli.POST("/api/save", ctrl.Save)
		webcli.POST("/api/saveendpoint", ctrl.SaveEndpoint)
	}

	r.Use(APIMiddleware())

	r.NoRoute(func(c *gin.Context) {
		Broadcast(c)
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

var wsupgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	connID := r.URL.Query().Get("connectionId")
	if connID != "" {
		conn := Clients[connID]
		if conn != nil {
			return
		}
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	Clients[connID] = conn

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			delete(Clients, connID)
			break
		}
		conn.WriteMessage(t, msg)
	}
}
