package ginex

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rickone/athena/limiter"
	"github.com/rickone/athena/logger"
	"github.com/rickone/athena/metrics"
	"github.com/rickone/athena/redis"
	"github.com/sirupsen/logrus"
)

func getFullMethod(c *gin.Context) string {
	return fmt.Sprintf("%s%s", c.Request.Method, c.Request.URL.Path)
}

func BlockerMW() gin.HandlerFunc {
	return Wrap(func(c *gin.Context) (interface{}, error) {
		fullMethod := getFullMethod(c)
		err := redis.CheckBlock("api", fullMethod)
		if err != nil {
			c.Abort()
			return nil, err
		}
		return nil, nil
	})
}

func LimiterMW(limit float64, bucket int, getId func(c *gin.Context) string) gin.HandlerFunc {
	lm := limiter.NewLimiterManager(limit, bucket)
	return Wrap(func(c *gin.Context) (interface{}, error) {
		id := getId(c)
		fullMethod := getFullMethod(c)

		err := lm.Allow(fullMethod, id)
		if err != nil {
			c.Abort()
			return nil, err
		}

		return nil, nil
	})
}

func RequestIdMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader("Request-Id")
		if requestId == "" {
			requestId = uuid.New().String()
		}
		c.Set("Request-Id", requestId)
		c.Writer.Header().Set("Request-Id", requestId)
	}
}

func AccessLogMW() gin.HandlerFunc {
	service := os.Getenv("Service")

	return func(c *gin.Context) {
		c.Set("Logger", logger.NewEntry(c, map[string]interface{}{
			"request_id": c.GetString("Request-Id"),
			"client_ip":  c.ClientIP(),
			"service":    service,
			"method":     getFullMethod(c),
		}))

		start := time.Now()
		c.Next()
		latency := time.Now().Sub(start).Milliseconds()
		code := c.Writer.Status()

		fields := map[string]interface{}{
			"latency": latency,
			"code":    code,
		}

		if len(c.Errors) > 0 {
			fields["err"] = c.Errors.ByType(gin.ErrorTypePrivate).String()
		}

		if code >= 200 && code < 400 {
			GetLogger(c).WithFields(fields).Info("Access success")
		} else if code >= 400 && code < 500 {
			GetLogger(c).WithFields(fields).Warn("Access denied")
		} else {
			GetLogger(c).WithFields(fields).Error("Access failed")
		}
	}
}

func GetLogger(c *gin.Context) *logrus.Entry {
	return c.MustGet("Logger").(*logrus.Entry)
}

func RecoveryMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if ret := recover(); ret != nil {
				GetLogger(c).WithFields(logrus.Fields{
					"stack": string(debug.Stack()),
					"err":   ret,
				}).Error("Recover panic")

				c.AbortWithStatus(http.StatusServiceUnavailable)
			}
		}()
		c.Next()
	}
}

func MetricsMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullMethod := getFullMethod(c)

		ts := time.Now()
		latency := metrics.NewHistogram("latency", "method", fullMethod)
		c.Next()
		latency.Update(time.Since(ts).Nanoseconds())

		status := "success"
		if c.Writer.Status() >= 500 {
			status = "failed"
		}

		call := metrics.NewCounter("call", "method", fullMethod, "status", status, "code", strconv.Itoa(c.Writer.Status()))
		call.Inc(1)
	}
}
