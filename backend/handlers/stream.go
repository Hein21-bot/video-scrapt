package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const chromeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func HandleStreamHEAD(c *gin.Context) {
	if c.Query("url") == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
}

func HandleStream(c *gin.Context) {
	rawURL  := c.Query("url")
	referer := c.Query("referer")
	if rawURL == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	isM3U8 := strings.Contains(rawURL, ".m3u8")

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
		if p, err := url.Parse(referer); err == nil {
			req.Header.Set("Origin", p.Scheme+"://"+p.Host)
		}
	}
	if !isM3U8 {
		if rangeH := c.GetHeader("Range"); rangeH != "" {
			req.Header.Set("Range", rangeH)
		}
	}

	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if referer != "" {
				r.Header.Set("Referer", referer)
			}
			r.Header.Set("User-Agent", chromeUA)
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !isM3U8 {
		isM3U8 = strings.Contains(ct, "mpegurl")
	}

	if isM3U8 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.Status(http.StatusBadGateway)
			return
		}
		rewritten := []byte(rewriteM3U8(string(body), rawURL, referer))
		c.Data(http.StatusOK, "application/vnd.apple.mpegurl", rewritten)
		return
	}

	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Cache-Control"} {
		if v := resp.Header.Get(h); v != "" {
			c.Header(h, v)
		}
	}
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func rewriteM3U8(content, baseURL, referer string) string {
	base, _ := url.Parse(baseURL)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		segURL := trimmed
		if !strings.HasPrefix(segURL, "http") {
			if ref, err := url.Parse(segURL); err == nil {
				segURL = base.ResolveReference(ref).String()
			}
		}
		lines[i] = "/api/stream?url=" + url.QueryEscape(segURL) + "&referer=" + url.QueryEscape(referer)
	}
	return strings.Join(lines, "\n")
}
