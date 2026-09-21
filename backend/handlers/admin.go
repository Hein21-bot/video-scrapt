package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/config"
	"video-scraper/db"
	"video-scraper/models"
	"video-scraper/scraper"
)

// ── Login ─────────────────────────────────────────────────────────────────────

func HandleAdminLogin(c *gin.Context) {
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}
	if body.Password != config.C.AdminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(config.C.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": signed})
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func HandleAdminStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, _ := db.ListingCol.CountDocuments(ctx, bson.M{})

	pipeline := bson.A{
		bson.M{"$group": bson.M{"_id": "$site", "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"count": -1}},
	}
	cur, err := db.ListingCol.Aggregate(ctx, pipeline)
	bySite := []gin.H{}
	if err == nil {
		var rows []struct {
			Site  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		cur.All(ctx, &rows)
		for _, r := range rows {
			bySite = append(bySite, gin.H{"site": r.Site, "count": r.Count})
		}
	}

	cacheCount, _ := db.CacheCol.CountDocuments(ctx, bson.M{})
	shareCount, _  := db.ShareCol.CountDocuments(ctx, bson.M{})

	c.JSON(http.StatusOK, gin.H{
		"total_videos":  total,
		"by_site":       bySite,
		"cache_entries": cacheCount,
		"share_links":   shareCount,
		"sync_running":  atomic.LoadInt32(&syncRunning) == 1,
		"last_sync":     lastSyncTime,
	})
}

// ── Videos ────────────────────────────────────────────────────────────────────

func HandleAdminVideos(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	skip := int64((page - 1) * 20)

	filter := bson.M{}
	if site := c.Query("site"); site != "" {
		filter["site"] = site
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		filter["title"] = bson.M{"$regex": q, "$options": "i"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	total, _ := db.ListingCol.CountDocuments(ctx, filter)

	opts := options.Find().
		SetSort(bson.D{{Key: "synced_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(20)

	cursor, err := db.ListingCol.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var docs []models.ListingDoc
	cursor.All(ctx, &docs)

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"items": docs,
	})
}

// HandleAdminCreateVideo adds a manually-entered video (the "Myanmar" channel).
// splitCSV turns "ai xi, yuna" into ["ai-xi","yuna"] (slug form, deduped).
func splitCSV(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, part := range strings.Split(s, ",") {
		slug := strings.ToLower(strings.Join(strings.Fields(part), "-"))
		if slug != "" && !seen[slug] {
			seen[slug] = true
			out = append(out, slug)
		}
	}
	return out
}

// It stores a directly-playable URL so HandleVideoURL can skip scraping.
func HandleAdminCreateVideo(c *gin.Context) {
	var body struct {
		Title     string `json:"title"`
		VideoURL  string `json:"video_url"`
		MirrorURL string `json:"mirror_url"`
		Thumbnail string `json:"thumbnail"`
		Channel   string `json:"channel"`  // which manual channel; defaults to Myanmar for the bridge
		Category  string `json:"category"` // sub-tab within the channel, e.g. language
		Actors    string `json:"actors"`   // comma-separated, optional
		Tags      string `json:"tags"`     // comma-separated, optional
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.VideoURL = strings.TrimSpace(body.VideoURL)
	body.MirrorURL = strings.TrimSpace(body.MirrorURL)
	if body.Title == "" || body.VideoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and video_url are required"})
		return
	}
	if !strings.HasPrefix(body.VideoURL, "http") ||
		(body.MirrorURL != "" && !strings.HasPrefix(body.MirrorURL, "http")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URLs must be http(s) links"})
		return
	}

	channel := strings.TrimSpace(body.Channel)
	if channel == "" {
		channel = "channel1"
	}
	site, ok := toInternalSite(channel)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel"})
		return
	}

	thumb := strings.TrimSpace(body.Thumbnail)
	if thumb == "" && scraper.EmbedHostRe.MatchString(body.VideoURL) {
		if r, err := scraper.ScrapeEmbed(body.VideoURL); err == nil {
			thumb = r.Thumb
		}
	}

	category := strings.ToLower(strings.TrimSpace(body.Category))
	if category == "" {
		category = "other"
	}

	doc := models.ListingDoc{
		PageURL:   "manual://" + models.RandCode(10),
		Title:     body.Title,
		Thumbnail: thumb,
		Site:      site,
		Category:  category,
		Actors:    splitCSV(body.Actors),
		Tags:      splitCSV(body.Tags),
		VideoURL:  body.VideoURL,
		MirrorURL: body.MirrorURL,
		Manual:    true,
		SyncedAt:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := db.ListingCol.InsertOne(ctx, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "added", "id": res.InsertedID})
}

func HandleAdminDeleteVideo(c *gin.Context) {
	var body struct {
		PageURL string `json:"page_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_url required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.ListingCol.DeleteOne(ctx, bson.M{"page_url": body.PageURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	// Also clear from cache
	db.CacheCol.DeleteOne(ctx, bson.M{"_id": body.PageURL})

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── Shares ────────────────────────────────────────────────────────────────────

func HandleAdminShares(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.ShareCol.Find(ctx, bson.M{}, options.Find().SetLimit(100).SetSort(bson.D{{Key: "_id", Value: -1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var docs []models.ShareDoc
	cursor.All(ctx, &docs)
	c.JSON(http.StatusOK, docs)
}
