package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChannelDoc is an admin-created manual channel (a main tab like "Myanmar", but
// user-defined). Key doubles as the internal ListingDoc.Site value for videos
// filed under it — unlike the two built-in channels (channel1/channel2), custom
// channels have no separate public-vs-internal name.
type ChannelDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Key       string             `bson:"key"`
	Label     string             `bson:"label"`
	Order     int                `bson:"order"`
	CreatedAt time.Time          `bson:"created_at"`
}
