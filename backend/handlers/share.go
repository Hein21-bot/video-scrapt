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

func HandleShareCreate(c *gin.Context) {
	var body struct {
		VideoID   string `json:"video_id"`
		Title     string `json:"title"`
		Thumbnail string `json:"thumbnail"`
		Site      string `json:"site"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.VideoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_id required"})
		return
	}

	oid, err := primitive.ObjectIDFromHex(body.VideoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video_id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var listing models.ListingDoc
	if err := db.ListingCol.FindOne(ctx, bson.M{"_id": oid}).Decode(&listing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	doc := models.ShareDoc{
		Code:      models.RandCode(7),
		VideoID:   body.VideoID,
		PageURL:   listing.PageURL,
		Title:     body.Title,
		Thumbnail: body.Thumbnail,
		Site:      body.Site,
	}

	db.ShareCol.InsertOne(ctx, doc)
	c.JSON(http.StatusOK, gin.H{"code": doc.Code})
}

func HandleShareGet(c *gin.Context) {
	code := c.Param("code")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc models.ShareDoc
	if err := db.ShareCol.FindOne(ctx, bson.M{"code": code}).Decode(&doc); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}
