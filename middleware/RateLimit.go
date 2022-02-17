package middleware

import (
	"camp/types"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"net/http"
	"time"
)

func RateLimitMiddleware(fillInterval time.Duration, cap int64, quantum int64) func(c *gin.Context) {
	// 创建指定填充速率、容量大小和每次填充的令牌数的令牌桶
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)
	return func(c *gin.Context) {
		// 如果取不到令牌就中断本次请求返回 rate limit...
		if bucket.TakeAvailable(1) < 1 {
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
			c.Abort()
			return
		}
		c.Next()
	}
}
