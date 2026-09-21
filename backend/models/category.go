package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CategoryDoc is an admin-managed sub-tab (e.g. the Myanmar channel's
// Japanese/English/Chinese language tabs). Value is the slug stored on
// ListingDoc.Category; Label is the display text shown on the tab.
type CategoryDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Site      string             `bson:"site"` // internal site key, e.g. "manual"
	Value     string             `bson:"value"`
	Label     string             `bson:"label"`
	Order     int                `bson:"order"`
	CreatedAt time.Time          `bson:"created_at"`
}
