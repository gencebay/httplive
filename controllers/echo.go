package controllers

import (
	"github.com/gin-gonic/gin"
)

type EchoController struct{}

func (ctrl EchoController) Echo(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
