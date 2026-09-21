package handlers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"video-scraper/db"
	"video-scraper/models"
)

var channelToSite = map[string]string{
	"channel1": "manual",
	"channel2": "3xchina",
	"channel3": "muskuduu",
}

var siteToChannel = map[string]string{
	"manual":   "channel1",
	"3xchina":  "channel2",
	"muskuduu": "channel3",
}

// reservedChannelKeys blocks admin-created channels from colliding with the
// built-in ones (which live in the maps above, not the channels collection).
var reservedChannelKeys = map[string]bool{
	"channel1": true, "channel2": true, "channel3": true,
	"manual": true, "3xchina": true, "muskuduu": true,
}

// toInternalSite maps a public channel name to the internal DB site name.
// Returns ("", false) for unknown channels so callers can reject the request.
// The two built-in channels are free; anything else is looked up in the
// admin-managed channels collection (a custom channel's key IS its site value).
func toInternalSite(channel string) (string, bool) {
	if s, ok := channelToSite[channel]; ok {
		return s, true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var doc models.ChannelDoc
	if err := db.ChannelCol.FindOne(ctx, bson.M{"key": channel}).Decode(&doc); err == nil {
		return doc.Key, true
	}
	return "", false
}

func toChannelName(site string) string {
	if c, ok := siteToChannel[site]; ok {
		return c
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var doc models.ChannelDoc
	if err := db.ChannelCol.FindOne(ctx, bson.M{"key": site}).Decode(&doc); err == nil {
		return doc.Key
	}
	return site
}
