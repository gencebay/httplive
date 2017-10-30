package controllers

import (
	"github.com/gin-gonic/gin"
)

// EchoController ...
type EchoController struct {
}

// Echo ...
func (ctrl EchoController) Echo(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
