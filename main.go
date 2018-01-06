package main

import (
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

	r.Use(CORSMiddleware())

	r.Use(ConfigJsMiddleware())

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

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

	if hasMultiplePort {
		for i := 1; i < length; i++ {
			go func(port string) {
				r.Run(":" + port)
			}(portsArr[i])
		}
	}

	r.Run(":" + port)
}
