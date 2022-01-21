package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetMember(c *gin.Context) {
	c.String(http.StatusOK, "pong, method is GET")
}
