package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// baseModel contains common fields for all MongoDB documents.
type baseModel struct {
	ID        *bson.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time      `bson:"created_at"`
	UpdatedAt time.Time      `bson:"updated_at"`
	DeletedAt *time.Time     `bson:"deleted_at,omitempty"`
}
