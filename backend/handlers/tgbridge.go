package handlers

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"video-scraper/config"
)

var bridgeClient = &http.Client{Timeout: 20 * time.Second}

// HandleBridgeStatus tells the admin UI whether the local Telegram bridge is up.
func HandleBridgeStatus(c *gin.Context) {
	if config.C.BridgeURL == "" {
		c.JSON(http.StatusOK, gin.H{"enabled": false})
		return
	}
	resp, err := bridgeClient.Get(config.C.BridgeURL + "/health")
	online := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}
	c.JSON(http.StatusOK, gin.H{"enabled": true, "online": online})
}

// HandleBridgeAdd / HandleBridgeJobs proxy to the local bridge (bridge/server.py).
// They only work when BRIDGE_URL is set — i.e. an admin running things locally.
func HandleBridgeAdd(c *gin.Context)  { proxyBridge(c, "POST", "/add") }
func HandleBridgeJobs(c *gin.Context) { proxyBridge(c, "GET", "/jobs") }

func proxyBridge(c *gin.Context, method, path string) {
	if config.C.BridgeURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "bridge not configured"})
		return
	}
	var body io.Reader
	if method != http.MethodGet {
		b, _ := io.ReadAll(c.Request.Body)
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, config.C.BridgeURL+path, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bridgeClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "bridge offline — start bridge/server.py"})
		return
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", out)
}
