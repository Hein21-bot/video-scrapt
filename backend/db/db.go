package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"video-scraper/config"
	"video-scraper/models"
)

var (
	Client          *mongo.Client
	CacheCol        *mongo.Collection
	ListingCol      *mongo.Collection
	ShareCol        *mongo.Collection
	CategoryCol     *mongo.Collection
	ChannelCol      *mongo.Collection
	ChannelLabelCol *mongo.Collection
)

func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(config.C.MongoURI))
	if err != nil {
		log.Fatalf("[db] connect: %v", err)
	}
	if err = Client.Ping(ctx, nil); err != nil {
		log.Fatalf("[db] ping: %v", err)
	}

	database := Client.Database(config.C.MongoDB)
	CacheCol    = database.Collection("video_cache")
	ListingCol  = database.Collection("video_listings")
	ShareCol    = database.Collection("video_shares")
	CategoryCol     = database.Collection("categories")
	ChannelCol      = database.Collection("channels")
	ChannelLabelCol = database.Collection("channel_labels")

	CacheCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	ListingCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "page_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	ListingCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "site", Value: 1},
			{Key: "category", Value: 1},
			{Key: "synced_at", Value: -1},
		},
	})
	ShareCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	CategoryCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "site", Value: 1}, {Key: "value", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	ChannelCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	if n, _ := CategoryCol.CountDocuments(ctx, bson.M{"site": "manual"}); n == 0 {
		defaults := []interface{}{
			models.CategoryDoc{Site: "manual", Value: "japanese", Label: "Japanese", Order: 0, CreatedAt: time.Now()},
			models.CategoryDoc{Site: "manual", Value: "english", Label: "English", Order: 1, CreatedAt: time.Now()},
			models.CategoryDoc{Site: "manual", Value: "chinese", Label: "Chinese", Order: 2, CreatedAt: time.Now()},
		}
		if _, err := CategoryCol.InsertMany(ctx, defaults); err != nil {
			log.Printf("[db] seed categories: %v", err)
		}
	}

	log.Printf("[db] connected → %s / %s", config.C.MongoURI, config.C.MongoDB)
}
