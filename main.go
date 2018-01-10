package main

import (
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

	Environments.DefaultPort = port
	Environments.HasMultiplePort = hasMultiplePort
	Environments.DatabaseAttachedFullPath = dbPath

	createDb()

	InitDbValues()

	r := gin.Default()

	r.Use(StaticFileMiddleware())

	r.Use(CORSMiddleware())

	r.Use(ConfigJsMiddleware())

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

	r.Use(APIMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.Status(404)
		c.File("./public/404.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		handleConnections(c.Writer, c.Request)
	})

	go HandleMessages()

	if hasMultiplePort {
		for i := 1; i < length; i++ {
			go func(port string) {
				r.Run(":" + port)
			}(portsArr[i])
		}
	}

	r.Run(":" + port)
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// get default ws upgrader value
	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// new client
	Clients[ws] = true

	for {
		var msg WsMessage

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(Clients, ws)
			break
		}
		// send new message to client
		Broadcast <- msg
	}
}
