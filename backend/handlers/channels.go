package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/db"
	"video-scraper/models"
)

// builtinChannels are hardcoded, not stored in the channels collection —
// they map to real code paths (an auto-scraper, the manual-video system).
// Their key/site can't change, but the display label can — see builtinLabel.
var builtinChannels = []struct{ Key, Label, Site string }{
	{"channel2", "Chinese AV", "3xchina"},
	{"channel1", "Myanmar", "manual"},
	{"channel3", "Muskuduu", "muskuduu"},
}

// builtinLabel returns the admin-set override for a builtin channel's label,
// falling back to the hardcoded default when none is set.
func builtinLabel(ctx context.Context, key, fallback string) string {
	var doc struct {
		Label string `bson:"label"`
	}
	if err := db.ChannelLabelCol.FindOne(ctx, bson.M{"_id": key}).Decode(&doc); err == nil && doc.Label != "" {
		return doc.Label
	}
	return fallback
}

// HandleChannelsList is public — powers the main tab bar on the user site.
func HandleChannelsList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := []gin.H{}
	for _, b := range builtinChannels {
		out = append(out, gin.H{"key": b.Key, "label": builtinLabel(ctx, b.Key, b.Label)})
	}

	cur, err := db.ChannelCol.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "order", Value: 1}}))
	if err == nil {
		var docs []models.ChannelDoc
		cur.All(ctx, &docs)
		for _, d := range docs {
			out = append(out, gin.H{"key": d.Key, "label": d.Label})
		}
	}
	c.JSON(http.StatusOK, out)
}

// HandleAdminChannelsList adds ids, video counts, and a builtin flag for the
// admin CRUD table. Builtins use their key as "id" (renaming a builtin only
// changes its label — see HandleAdminChannelUpdate — so no ObjectID exists).
func HandleAdminChannelsList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := []gin.H{}
	for _, b := range builtinChannels {
		count, _ := db.ListingCol.CountDocuments(ctx, bson.M{"site": b.Site})
		label := builtinLabel(ctx, b.Key, b.Label)
		out = append(out, gin.H{"id": b.Key, "key": b.Key, "label": label, "count": count, "builtin": true})
	}

	cur, err := db.ChannelCol.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "order", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var docs []models.ChannelDoc
	cur.All(ctx, &docs)
	for _, d := range docs {
		count, _ := db.ListingCol.CountDocuments(ctx, bson.M{"site": d.Key})
		out = append(out, gin.H{"id": d.ID.Hex(), "key": d.Key, "label": d.Label, "count": count, "builtin": false})
	}
	c.JSON(http.StatusOK, out)
}

func HandleAdminChannelCreate(c *gin.Context) {
	var body struct {
		Label string `json:"label"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	label := strings.TrimSpace(body.Label)
	key := slugifyLabel(label)
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}
	if reservedChannelKeys[key] {
		c.JSON(http.StatusConflict, gin.H{"error": "that name is reserved"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, _ := db.ChannelCol.CountDocuments(ctx, bson.M{})
	doc := models.ChannelDoc{Key: key, Label: label, Order: int(count), CreatedAt: time.Now()}

	res, err := db.ChannelCol.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "a channel with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": res.InsertedID.(primitive.ObjectID).Hex(), "key": key, "label": label})
}

func HandleAdminChannelUpdate(c *gin.Context) {
	idParam := c.Param("id")
	var body struct {
		Label string `json:"label"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	label := strings.TrimSpace(body.Label)
	if label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Builtin channels keep their key (it's wired to real scraper/manual-video
	// code) — only the display label can change, stored as an override.
	for _, b := range builtinChannels {
		if b.Key == idParam {
			_, err := db.ChannelLabelCol.UpdateByID(ctx, idParam,
				bson.M{"$set": bson.M{"label": label}}, options.Update().SetUpsert(true))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "updated", "key": b.Key, "label": label})
			return
		}
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	newKey := slugifyLabel(label)
	if newKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}
	if reservedChannelKeys[newKey] {
		c.JSON(http.StatusConflict, gin.H{"error": "that name is reserved"})
		return
	}

	var existing models.ChannelDoc
	if err := db.ChannelCol.FindOne(ctx, bson.M{"_id": id}).Decode(&existing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if _, err := db.ChannelCol.UpdateByID(ctx, id, bson.M{"$set": bson.M{"label": label, "key": newKey}}); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "a channel with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Renaming changes the key — repoint its videos and sub-tabs.
	if newKey != existing.Key {
		db.ListingCol.UpdateMany(ctx, bson.M{"site": existing.Key}, bson.M{"$set": bson.M{"site": newKey}})
		db.CategoryCol.UpdateMany(ctx, bson.M{"site": existing.Key}, bson.M{"$set": bson.M{"site": newKey}})
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated", "key": newKey, "label": label})
}

func HandleAdminChannelDelete(c *gin.Context) {
	idParam := c.Param("id")
	for _, b := range builtinChannels {
		if b.Key == idParam {
			c.JSON(http.StatusConflict, gin.H{"error": "built-in channels can't be deleted"})
			return
		}
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing models.ChannelDoc
	if err := db.ChannelCol.FindOne(ctx, bson.M{"_id": id}).Decode(&existing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if n, _ := db.ListingCol.CountDocuments(ctx, bson.M{"site": existing.Key}); n > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "this channel still has videos — delete or move them first"})
		return
	}

	if _, err := db.ChannelCol.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.CategoryCol.DeleteMany(ctx, bson.M{"site": existing.Key})

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
