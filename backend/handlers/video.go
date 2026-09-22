package handlers

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/db"
	"video-scraper/models"
	"video-scraper/scraper"
)

const cacheTTL = 2 * time.Hour

func BuildProxyURL(videoURL, referer string) string {
	return "/api/stream?url=" + url.QueryEscape(videoURL) + "&referer=" + url.QueryEscape(referer)
}

func cacheGet(pageURL string) (*scraper.VideoResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var doc models.CacheDoc
	err := db.CacheCol.FindOne(ctx, bson.M{
		"_id":        pageURL,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&doc)
	if err != nil {
		return nil, false
	}
	return &scraper.VideoResult{URL: doc.URL, Type: doc.Type, Mirrors: doc.Mirrors}, true
}

func cacheSet(pageURL string, r *scraper.VideoResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.CacheCol.ReplaceOne(ctx,
		bson.M{"_id": pageURL},
		models.CacheDoc{
			PageURL:   pageURL,
			URL:       r.URL,
			Type:      r.Type,
			Mirrors:   r.Mirrors,
			ExpiresAt: time.Now().Add(cacheTTL),
		},
		options.Replace().SetUpsert(true),
	)
}

func HandleVideoList(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	skip := int64((page - 1) * 20)

	filter := bson.M{}
	if site := c.Query("site"); site != "" {
		internal, ok := toInternalSite(site)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
			return
		}
		filter["site"] = internal
	}
	if category := c.Query("category"); category != "" {
		filter["category"] = category
	}
	if actor := strings.TrimSpace(c.Query("actor")); actor != "" {
		filter["actors"] = actor
	}
	if tag := strings.TrimSpace(c.Query("tag")); tag != "" {
		// Some channels (Muskuduu) have no per-video tags, only a genre
		// (categories) — match either so its genre chips can filter too.
		filter["$or"] = bson.A{
			bson.M{"tags": tag},
			bson.M{"categories": tag},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	total, err := db.ListingCol.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	if err := cursor.All(ctx, &docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cards := make([]scraper.VideoCard, len(docs))
	for i, d := range docs {
		cards[i] = scraper.VideoCard{
			ID:        d.ID.Hex(),
			Title:     d.Title,
			Thumbnail: ProxyThumb(d.ID.Hex()),
			Site:      toChannelName(d.Site),
			SyncedAt:  d.SyncedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": cards, "total": total})
}

func HandleSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Match the title, the slug/code in the page URL, an actress name, or a tag/genre —
	// spaces and dashes are interchangeable so "xjx 720" finds "XJX-720".
	parts := regexp.MustCompile(`[\s-]+`).Split(q, -1)
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	rx := strings.Join(parts, `[\s-]?`)
	filter := bson.M{"$or": bson.A{
		bson.M{"title": bson.M{"$regex": rx, "$options": "i"}},
		bson.M{"page_url": bson.M{"$regex": rx, "$options": "i"}},
		bson.M{"actors": bson.M{"$regex": rx, "$options": "i"}},
		bson.M{"tags": bson.M{"$regex": rx, "$options": "i"}},
		bson.M{"categories": bson.M{"$regex": rx, "$options": "i"}},
	}}
	// Optional: keep results inside the channel the visitor is browsing.
	if site := c.Query("site"); site != "" {
		internal, ok := toInternalSite(site)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
			return
		}
		filter["site"] = internal
	}
	cursor, err := db.ListingCol.Find(ctx,
		filter,
		options.Find().SetLimit(40).SetSort(bson.D{{Key: "synced_at", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var docs []models.ListingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cards := make([]scraper.VideoCard, len(docs))
	for i, d := range docs {
		cards[i] = scraper.VideoCard{ID: d.ID.Hex(), Title: d.Title, Thumbnail: ProxyThumb(d.ID.Hex()), Site: toChannelName(d.Site)}
	}
	c.JSON(http.StatusOK, cards)
}

// HandleActors returns the list of actresses that appear on stored videos,
// ordered by how many videos each has. Used to populate the filter dropdown.
func HandleActors(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	match := bson.M{"actors.0": bson.M{"$exists": true}}
	if site := c.Query("site"); site != "" {
		internal, ok := toInternalSite(site)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
			return
		}
		match["site"] = internal
	}

	cur, err := db.ListingCol.Aggregate(ctx, bson.A{
		bson.M{"$match": match},
		bson.M{"$unwind": "$actors"},
		bson.M{"$sort": bson.M{"synced_at": -1}},
		bson.M{"$group": bson.M{
			"_id":   "$actors",
			"count": bson.M{"$sum": 1},
			"vid":   bson.M{"$first": "$_id"}, // newest video → its thumbnail
		}},
		bson.M{"$sort": bson.D{{Key: "count", Value: -1}, {Key: "_id", Value: 1}}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var rows []struct {
		Slug  string             `bson:"_id"`
		Count int                `bson:"count"`
		Vid   primitive.ObjectID `bson:"vid"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"slug":  r.Slug,
			"name":  prettyName(r.Slug),
			"count": r.Count,
			"thumb": ProxyThumb(r.Vid.Hex()),
		})
	}
	c.JSON(http.StatusOK, out)
}

var acronyms = map[string]bool{"av": true, "ntr": true, "pov": true, "bdsm": true, "jav": true, "milf": true, "3d": true, "sm": true}
var hasDigit = regexp.MustCompile(`\d`)

// prettyName title-cases an actress slug: "xin-yuri" → "Xin Yuri".
func prettyName(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

var allDigits = regexp.MustCompile(`^\d+$`)

// prettyLabel formats a tag/category slug, upper-casing codes and acronyms:
// "china-av" → "China AV" · "md-0334" → "MD 0334" · "ttp-008" → "TTP 008" · "3p" → "3P".
func prettyLabel(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		next := ""
		if i+1 < len(words) {
			next = words[i+1]
		}
		switch {
		case w == "":
		case len(w) <= 2, acronyms[w], hasDigit.MatchString(w), allDigits.MatchString(next):
			words[i] = strings.ToUpper(w)
		default:
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func HandleRelated(c *gin.Context) {
	site      := strings.TrimSpace(c.Query("site"))
	excludeID := c.Query("exclude")
	if site == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	internal, ok := toInternalSite(site)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
		return
	}
	filter := bson.M{"site": internal}
	if oid, err := primitive.ObjectIDFromHex(excludeID); err == nil {
		filter["_id"] = bson.M{"$ne": oid}
	}

	total, _ := db.ListingCol.CountDocuments(ctx, filter)
	skip := int64(0)
	if total > 12 {
		skip = rand.Int63n(total - 12)
	}

	cursor, err := db.ListingCol.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(12))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var docs []models.ListingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cards := make([]scraper.VideoCard, len(docs))
	for i, d := range docs {
		cards[i] = scraper.VideoCard{ID: d.ID.Hex(), Title: d.Title, Thumbnail: ProxyThumb(d.ID.Hex()), Site: toChannelName(d.Site)}
	}
	c.JSON(http.StatusOK, cards)
}

func HandleVideoURL(c *gin.Context) {
	oid, err := primitive.ObjectIDFromHex(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Look up the real source URL — never exposed to the client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var doc models.ListingDoc
	if err := db.ListingCol.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	pageURL := doc.PageURL

	// Manually-added videos carry the URL(s) directly.
	if doc.Manual {
		url, typ, thumb := resolveManualURL(doc.VideoURL)
		mirrors := []string{url}
		if doc.MirrorURL != "" {
			if mu, mt, _ := resolveManualURL(doc.MirrorURL); mt == typ {
				mirrors = append(mirrors, mu)
			}
		}
		// Backfill a poster the first time we manage to scrape one.
		if doc.Thumbnail == "" && thumb != "" {
			db.ListingCol.UpdateByID(ctx, oid, bson.M{"$set": bson.M{"thumbnail": thumb}})
		}
		c.JSON(http.StatusOK, gin.H{
			"url":        url,
			"type":       typ,
			"mirrors":    mirrors,
			"title":      doc.Title,
			"categories": categoryLabel(doc),
			"actors":     actorRefs(doc.Actors),
			"tags":       tagChips(doc),
		})
		return
	}

	result, ok := cacheGet(pageURL)
	if !ok {
		scraped, err := scraper.ScrapeVideoURL(pageURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		// Both supported CDNs serve their playlists and segments directly to the
		// browser with permissive CORS and no Referer lock, so the URL is handed
		// to the client as-is — no proxy, no server bandwidth.
		cacheSet(pageURL, scraped)
		result = scraped
	}

	c.JSON(http.StatusOK, gin.H{
		"url":        result.URL,
		"type":       result.Type,
		"mirrors":    result.Mirrors,
		"title":      doc.Title,
		"categories": categoryLabel(doc),
		"actors":     actorRefs(doc.Actors),
		"tags":       tagChips(doc),
	})
}

// resolveManualURL turns a stored manual URL into something the player can use:
//   - a direct .m3u8 / .mp4  → played as-is
//   - a known video-host embed → the extracted m3u8, or the embed page in an
//     <iframe> if extraction fails (host changed its obfuscation)
//   - anything else            → an <iframe>
func resolveManualURL(raw string) (playURL, playType, thumb string) {
	raw = strings.TrimSpace(raw)
	low := strings.ToLower(raw)
	switch {
	case strings.Contains(low, ".m3u8"):
		return raw, "hls", ""
	case strings.Contains(low, ".mp4"):
		return raw, "mp4", ""
	case scraper.EmbedHostRe.MatchString(raw):
		if scraper.IframeMode() {
			return scraper.NormalizeEmbedURL(raw), "iframe", ""
		}
		if r, err := scraper.ScrapeEmbed(raw); err == nil {
			return r.URL, "hls", r.Thumb
		}
		return scraper.NormalizeEmbedURL(raw), "iframe", "" // fall back to the host's own player
	default:
		return raw, "iframe", ""
	}
}

// actorRefs pairs each actress slug with a display name for clickable chips.
func actorRefs(slugs []string) []gin.H {
	out := make([]gin.H, 0, len(slugs))
	for _, s := range slugs {
		out = append(out, gin.H{"slug": s, "name": prettyName(s)})
	}
	return out
}

// prettyList turns tag/category slugs into {slug, label} pairs, keeping codes and
// acronyms upper-case ("china-av" → "China AV", "md-0334" → "MD 0334").
func prettyList(slugs []string) []gin.H {
	out := make([]gin.H, 0, len(slugs))
	for _, s := range slugs {
		out = append(out, gin.H{"slug": s, "label": prettyLabel(s)})
	}
	return out
}

// tagChips returns the clickable filter chips shown under a video. Most sites
// (3xchina) have real per-video tags; Muskuduu only has a genre (Categories),
// so that's promoted to fill the same role there.
func tagChips(doc models.ListingDoc) []gin.H {
	if len(doc.Tags) > 0 {
		return prettyList(doc.Tags)
	}
	return prettyList(doc.Categories)
}

// categoryLabel is the plain (non-clickable) genre line. When Categories was
// already promoted into tagChips above, it's left out here to avoid showing
// the same thing twice.
func categoryLabel(doc models.ListingDoc) []gin.H {
	if len(doc.Tags) > 0 {
		return prettyList(doc.Categories)
	}
	return nil
}
