package test

import (
	"camp/routers"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMember(t *testing.T) {

	r := gin.Default()
	routers.RegisterRouter(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/member/create", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
