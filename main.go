package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

func createDb(dbPath string, dbPathPresent bool) error {
	var err error

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		var filename string = filepath.Base(dbPath)
		if filename == DefaultDbName && !dbPathPresent {
			_, err := os.Create(dbPath)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	Environments.DatabaseFullPath = dbPath

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

	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	Environments.WorkingDirectory = filepath.ToSlash(workdir)
	Environments.DefaultPort = port
	Environments.HasMultiplePort = hasMultiplePort

	dbPathPresent := false
	if dbPath == "" {
		dbPath = path.Join(workdir, DefaultDbName)
	} else {
		dbPathPresent = true
	}

	createDb(dbPath, dbPathPresent)

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

	fmt.Printf("Httplive started with port: %s", port)
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
