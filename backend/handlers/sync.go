package handlers

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/db"
	"video-scraper/models"
	"video-scraper/scraper"
)

var (
	syncRunning  int32
	lastSyncTime time.Time
)

func syncTarget(t models.SyncTarget) {
	now := time.Now()
	count := 0
	for pg := 1; ; pg++ {
		cards, err := scraper.ScrapeVideoList(t.URL, pg)
		if err != nil || len(cards) == 0 {
			break
		}
		var writeModels []mongo.WriteModel
		for i, card := range cards {
			set := bson.M{
				"title":     card.Title,
				"thumbnail": card.Thumbnail,
				"site":      t.Site,
				"category":  t.Category,
			}
			if len(card.Actors) > 0 {
				set["actors"] = card.Actors
			}
			if len(card.Tags) > 0 {
				set["tags"] = card.Tags
			}
			if len(card.Categories) > 0 {
				set["categories"] = card.Categories
			}
			// synced_at is only set when a video is first inserted, so re-syncing
			// doesn't make old videos look new. The per-position offset keeps the
			// source site's newest-first order when a whole site is imported at once.
			firstSeen := now.Add(-time.Duration(count+i) * time.Millisecond)
			writeModels = append(writeModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"page_url": card.PageURL}).
				SetUpdate(bson.M{"$set": set, "$setOnInsert": bson.M{"synced_at": firstSeen}}).
				SetUpsert(true))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		db.ListingCol.BulkWrite(ctx, writeModels, options.BulkWrite().SetOrdered(false))
		cancel()
		count += len(cards)
		if len(cards) < 20 {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	log.Printf("[sync] %s/%s: %d videos", t.Site, t.Category, count)
}

func SyncListings() {
	if !atomic.CompareAndSwapInt32(&syncRunning, 0, 1) {
		log.Println("[sync] already running, skipping")
		return
	}
	defer atomic.StoreInt32(&syncRunning, 0)

	log.Println("[sync] starting...")
	for _, t := range models.SyncTargets {
		syncTarget(t)
	}
	lastSyncTime = time.Now()
	log.Println("[sync] done")
}

func StartBackground() {
	go func() {
		for range time.Tick(6 * time.Hour) {
			SyncListings()
		}
	}()
}

func HandleSync(c *gin.Context) {
	if atomic.LoadInt32(&syncRunning) == 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "sync already running"})
		return
	}
	go SyncListings()
	c.JSON(http.StatusOK, gin.H{"message": "sync started"})
}

func HandleSyncTarget(c *gin.Context) {
	var body struct {
		Site     string `json:"site"`
		Category string `json:"category"`
		URL      string `json:"url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Site == "" || body.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site and url required"})
		return
	}
	if !atomic.CompareAndSwapInt32(&syncRunning, 0, 1) {
		c.JSON(http.StatusConflict, gin.H{"error": "sync already running"})
		return
	}
	go func() {
		defer atomic.StoreInt32(&syncRunning, 0)
		syncTarget(models.SyncTarget{Site: body.Site, Category: body.Category, URL: body.URL})
		lastSyncTime = time.Now()
	}()
	c.JSON(http.StatusOK, gin.H{"message": "syncing " + body.Site + "/" + body.Category})
}

func HandleSyncStatus(c *gin.Context) {
	running := atomic.LoadInt32(&syncRunning) == 1
	c.JSON(http.StatusOK, gin.H{
		"running":        running,
		"last_sync_time": lastSyncTime,
	})
}
