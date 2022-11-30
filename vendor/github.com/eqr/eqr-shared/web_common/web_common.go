package web_common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowErrorMessage(c *gin.Context, message string) {
	c.HTML(
		http.StatusInternalServerError,
		"error.html",
		gin.H{
			"message": message,
			"title":   "Streams",
		},
	)
}

func ShowError(c *gin.Context, err error) {
	ShowErrorMessage(c, err.Error())
}

func Redirect(c *gin.Context, targetUrl string) {
	c.Redirect(http.StatusFound, targetUrl)
	c.Abort()
}
