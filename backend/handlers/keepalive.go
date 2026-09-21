package handlers

import (
	"context"
	"crypto/subtle"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/config"
	"video-scraper/db"
	"video-scraper/models"
	"video-scraper/scraper"
)

// The free video hosts (StreamWish / VidHide) delete files that see no activity
// for months. "Pinning" fetches the first bytes of every manually-added video's
// stream — exactly what a player does — so the host counts it as watched.

var (
	pinRunning int32
	pinMu      sync.Mutex
	pinLast    struct {
		At                time.Time `json:"last_run"`
		Total, OK, Failed int
	}
	pinClient = &http.Client{Timeout: 20 * time.Second}
)

func authKeepAlive(c *gin.Context) bool {
	key := config.C.KeepAliveKey
	if key == "" {
		c.Status(http.StatusNotFound)
		return false
	}
	if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Keepalive-Key")), []byte(key)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	return true
}

// HandleKeepAlive starts a pin run in the background and returns immediately.
func HandleKeepAlive(c *gin.Context) {
	if !authKeepAlive(c) {
		return
	}
	if !atomic.CompareAndSwapInt32(&pinRunning, 0, 1) {
		c.JSON(http.StatusConflict, gin.H{"status": "already running"})
		return
	}
	go func() {
		defer atomic.StoreInt32(&pinRunning, 0)
		runPin()
	}()
	c.JSON(http.StatusAccepted, gin.H{"status": "started"})
}

func HandleKeepAliveStatus(c *gin.Context) {
	if !authKeepAlive(c) {
		return
	}
	pinMu.Lock()
	defer pinMu.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"running":  atomic.LoadInt32(&pinRunning) == 1,
		"last_run": pinLast.At,
		"total":    pinLast.Total,
		"ok":       pinLast.OK,
		"failed":   pinLast.Failed,
	})
}

func runPin() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := db.ListingCol.Find(ctx, bson.M{"manual": true},
		options.Find().SetProjection(bson.M{"video_url": 1, "mirror_url": 1}))
	if err != nil {
		log.Printf("[pin] query: %v", err)
		return
	}
	var docs []models.ListingDoc
	if err := cur.All(ctx, &docs); err != nil {
		log.Printf("[pin] read: %v", err)
		return
	}

	var jobs []string
	for _, d := range docs {
		for _, u := range []string{d.VideoURL, d.MirrorURL} {
			if strings.TrimSpace(u) != "" {
				jobs = append(jobs, u)
			}
		}
	}

	var ok, failed int32
	ch := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range ch {
				if err := pinURL(u); err != nil {
					atomic.AddInt32(&failed, 1)
					log.Printf("[pin] %s: %v", u, err)
				} else {
					atomic.AddInt32(&ok, 1)
				}
				time.Sleep(300 * time.Millisecond)
			}
		}()
	}
	for _, u := range jobs {
		ch <- u
	}
	close(ch)
	wg.Wait()

	pinMu.Lock()
	pinLast.At, pinLast.Total, pinLast.OK, pinLast.Failed = time.Now(), len(jobs), int(ok), int(failed)
	pinMu.Unlock()
	log.Printf("[pin] done: %d links, %d ok, %d failed", len(jobs), ok, failed)
}

// pinURL touches one stored video link the way a viewer's player would.
func pinURL(raw string) error {
	raw = strings.TrimSpace(raw)
	low := strings.ToLower(raw)
	switch {
	case strings.Contains(low, ".m3u8"):
		return pinHLS(raw, "")
	case strings.Contains(low, ".mp4"):
		_, err := fetchHead(raw, "", true)
		return err
	case scraper.EmbedHostRe.MatchString(raw):
		r, err := scraper.ScrapeEmbed(raw) // also loads the embed page itself
		if err != nil {
			return err
		}
		origin := raw
		if u, e := url.Parse(scraper.NormalizeEmbedURL(raw)); e == nil {
			origin = u.Scheme + "://" + u.Host + "/"
		}
		return pinHLS(r.URL, origin)
	default:
		_, err := fetchHead(raw, raw, false)
		return err
	}
}

// pinHLS loads the master playlist, then the first variant playlist, then the
// first bytes of the first media segment.
func pinHLS(playlist, referer string) error {
	cur := playlist
	for depth := 0; depth < 3; depth++ {
		body, err := fetchHead(cur, referer, false)
		if err != nil {
			return err
		}
		next := firstURI(string(body))
		if next == "" {
			return nil
		}
		base, err := url.Parse(cur)
		if err != nil {
			return err
		}
		ref, err := url.Parse(next)
		if err != nil {
			return err
		}
		cur = base.ResolveReference(ref).String()
		if !strings.Contains(strings.ToLower(cur), ".m3u8") {
			_, err := fetchHead(cur, referer, true) // a media segment
			return err
		}
	}
	return nil
}

func firstURI(playlist string) string {
	for _, line := range strings.Split(playlist, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

// fetchHead GETs a URL (only the first 2 KB when ranged) and returns what it read.
func fetchHead(u, referer string, ranged bool) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if ranged {
		req.Header.Set("Range", "bytes=0-2047")
	}
	resp, err := pinClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, &httpStatusError{resp.StatusCode}
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64*1024))
}

type httpStatusError struct{ code int }

func (e *httpStatusError) Error() string { return "status " + http.StatusText(e.code) }
