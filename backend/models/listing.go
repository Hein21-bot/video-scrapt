package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListingDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"         json:"-"`
	PageURL    string             `bson:"page_url"              json:"page_url"`
	Title      string             `bson:"title"                 json:"title"`
	Thumbnail  string             `bson:"thumbnail"             json:"thumbnail"`
	Site       string             `bson:"site"                  json:"site"`
	Category   string             `bson:"category"              json:"category"`
	Actors     []string           `bson:"actors,omitempty"      json:"-"`
	Tags       []string           `bson:"tags,omitempty"        json:"-"`
	Categories []string           `bson:"categories,omitempty"  json:"-"`
	// Manual entries carry the playable URL(s) directly (no page scraping).
	// VideoURL / MirrorURL may be a direct .mp4/.m3u8 or a video-host embed link.
	VideoURL  string    `bson:"video_url,omitempty"   json:"-"`
	MirrorURL string    `bson:"mirror_url,omitempty"  json:"-"`
	VideoType string    `bson:"video_type,omitempty"  json:"-"`
	Manual    bool      `bson:"manual,omitempty"      json:"-"`
	SyncedAt  time.Time `bson:"synced_at"             json:"synced_at"`
}

type SyncTarget struct {
	Site     string
	Category string
	URL      string
}

var SyncTargets = []SyncTarget{
	{Site: "3xchina", Category: "all", URL: "https://3xchina.page/"},
	{Site: "muskuduu", Category: "all", URL: "https://muskuduu.com/"},
}
