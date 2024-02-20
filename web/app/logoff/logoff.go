package logoff

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	token, err := c.Cookie("connection")
	if err != nil {
		c.SetCookie("connection", token, -500, "/", "http:localhost", true, true)
	}
	c.Redirect(http.StatusSeeOther, "/home")

}
