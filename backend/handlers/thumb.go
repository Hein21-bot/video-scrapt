package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"video-scraper/db"
	"video-scraper/models"
)

// HandleThumb resolves a video ID to its source thumbnail and redirects the
// browser straight there. Both supported sites (mgzaw, 3xchina) serve images
// with no Referer lock and permissive CORS, so proxying the bytes would only
// burn our bandwidth. The real image host still isn't in the API JSON — the
// client only learns it from the 302 Location on demand.
func HandleThumb(c *gin.Context) {
	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var doc models.ListingDoc
	if err := db.ListingCol.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if doc.Thumbnail == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Redirect(http.StatusFound, doc.Thumbnail)
}

func ProxyThumb(id string) string {
	if id == "" {
		return ""
	}
	return "/api/thumb/" + id
}
