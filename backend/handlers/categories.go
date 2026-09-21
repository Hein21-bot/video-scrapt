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

func slugifyLabel(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), "-"))
}

// HandleCategoriesList is public — powers the sub-tab bar on the user site.
func HandleCategoriesList(c *gin.Context) {
	internal, ok := toInternalSite(c.Query("site"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := db.CategoryCol.Find(ctx, bson.M{"site": internal}, options.Find().SetSort(bson.D{{Key: "order", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var docs []models.CategoryDoc
	cur.All(ctx, &docs)

	out := make([]gin.H, len(docs))
	for i, d := range docs {
		out[i] = gin.H{"value": d.Value, "label": d.Label}
	}
	c.JSON(http.StatusOK, out)
}

// HandleAdminCategoriesList is the same list with ids and per-tab video counts,
// for the admin CRUD table.
func HandleAdminCategoriesList(c *gin.Context) {
	internal, ok := toInternalSite(c.Query("site"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := db.CategoryCol.Find(ctx, bson.M{"site": internal}, options.Find().SetSort(bson.D{{Key: "order", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var docs []models.CategoryDoc
	cur.All(ctx, &docs)

	out := make([]gin.H, len(docs))
	for i, d := range docs {
		count, _ := db.ListingCol.CountDocuments(ctx, bson.M{"site": internal, "category": d.Value})
		out[i] = gin.H{"id": d.ID.Hex(), "value": d.Value, "label": d.Label, "count": count}
	}
	c.JSON(http.StatusOK, out)
}

func HandleAdminCategoryCreate(c *gin.Context) {
	var body struct {
		Site  string `json:"site"`
		Label string `json:"label"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	internal, ok := toInternalSite(body.Site)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site"})
		return
	}
	label := strings.TrimSpace(body.Label)
	value := slugifyLabel(label)
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, _ := db.CategoryCol.CountDocuments(ctx, bson.M{"site": internal})
	doc := models.CategoryDoc{Site: internal, Value: value, Label: label, Order: int(count), CreatedAt: time.Now()}

	res, err := db.CategoryCol.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "a sub-tab with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": res.InsertedID.(primitive.ObjectID).Hex(), "value": value, "label": label})
}

func HandleAdminCategoryUpdate(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Label string `json:"label"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	label := strings.TrimSpace(body.Label)
	newValue := slugifyLabel(label)
	if newValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing models.CategoryDoc
	if err := db.CategoryCol.FindOne(ctx, bson.M{"_id": id}).Decode(&existing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub-tab not found"})
		return
	}

	if _, err := db.CategoryCol.UpdateByID(ctx, id, bson.M{"$set": bson.M{"label": label, "value": newValue}}); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "a sub-tab with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Renaming changes the slug — repoint existing videos so they stay filed under it.
	if newValue != existing.Value {
		db.ListingCol.UpdateMany(ctx,
			bson.M{"site": existing.Site, "category": existing.Value},
			bson.M{"$set": bson.M{"category": newValue}},
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated", "value": newValue, "label": label})
}

func HandleAdminCategoryDelete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.CategoryCol.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub-tab not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
