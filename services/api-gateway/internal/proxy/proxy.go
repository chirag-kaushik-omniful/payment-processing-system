package proxy

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceProxy struct {
	client *http.Client
}

func NewServiceProxy() *ServiceProxy {
	return &ServiceProxy{client: &http.Client{Timeout: 30 * time.Second}}
}

func (p *ServiceProxy) Forward(c *gin.Context, baseURL, path string) {
	target := strings.TrimRight(baseURL, "/") + path
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "proxy error"})
		return
	}
	copyHeaders(c.Request.Header, req.Header)
	if uid := c.GetString("user_id"); uid != "" {
		req.Header.Set("X-User-ID", uid)
	}
	if cid, ok := c.Get("X-Correlation-ID"); ok {
		req.Header.Set("X-Correlation-ID", cid.(string))
	}
	if auth := c.GetHeader("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if key := c.GetHeader("Idempotency-Key"); key != "" {
		req.Header.Set("Idempotency-Key", key)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "upstream unavailable"})
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func copyHeaders(src, dst http.Header) {
	for k, vals := range src {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}
