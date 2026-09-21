package models

import "time"

type CacheDoc struct {
	PageURL   string    `bson:"_id"`
	URL       string    `bson:"url"`
	Type      string    `bson:"type"`
	Mirrors   []string  `bson:"mirrors"`
	ExpiresAt time.Time `bson:"expires_at"`
}
