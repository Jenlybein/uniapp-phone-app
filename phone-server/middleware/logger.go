package middleware

import (
	"bytes"
	"io"
	"time"

	"phone-server/utils"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter 用于记录响应体的包装器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 请求ID（如果没有则生成一个简单的）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = time.Now().Format("20060102150405") + "-" + c.ClientIP()
		}
		c.Set("request_id", requestID)

		// 请求方法、路径、查询参数
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 记录请求体（如果是POST/PUT等方法）
		var requestBody string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// 保存原始请求体
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 重置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 记录请求信息
		utils.Infofc(c.Request.Context(), "[REQUEST] request_id=%s, method=%s, path=%s, query=%s, client_ip=%s, body=%s",
			requestID, method, path, query, c.ClientIP(), requestBody)

		// 响应写入器包装
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// 记录响应信息
		statusCode := c.Writer.Status()
		responseBody := blw.body.String()
		bodySize := c.Writer.Size()

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			utils.Errorf("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s, response=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP(), responseBody)
		} else if statusCode >= 400 {
			utils.Warnf("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s, response=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP(), responseBody)
		} else {
			utils.Infof("[RESPONSE] request_id=%s, method=%s, path=%s, status=%d, duration=%v, body_size=%d, client_ip=%s",
				requestID, method, path, statusCode, duration, bodySize, c.ClientIP())
		}
	}
}