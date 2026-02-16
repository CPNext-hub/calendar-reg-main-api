package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const userCollection = "users"

// userModel is the MongoDB-specific representation of a user.
type userModel struct {
	BaseModel `bson:",inline"`
	Username  string `bson:"username"`
	Password  string `bson:"password"`
	Role      string `bson:"role"`
}

// toEntity converts a MongoDB model to a domain entity.
func (m *userModel) toEntity() *entity.User {
	var id string
	if m.ID != nil {
		id = m.ID.Hex()
	}

	return &entity.User{
		BaseEntity: entity.BaseEntity{
			ID:        id,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
		Username: m.Username,
		Password: m.Password,
		Role:     m.Role,
	}
}

// toUserModel converts a domain entity to a MongoDB model.
func toUserModel(e *entity.User) *userModel {
	m := &userModel{
		Username: e.Username,
		Password: e.Password,
		Role:     string(e.Role),
	}
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = e.DeletedAt
	if e.ID != "" {
		oid, err := bson.ObjectIDFromHex(e.ID)
		if err == nil {
			m.ID = &oid
		}
	}
	return m
}

type userRepository struct {
	db *mongo.Database
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *mongo.Database) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	model := toUserModel(user)
	result, err := r.db.Collection(userCollection).InsertOne(ctx, model)
	if err != nil {
		return err
	}

	// Write back the generated ID to the entity.
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		user.ID = oid.Hex()
	}
	return nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var model userModel
	filter := bson.M{"username": username, "deleted_at": bson.M{"$exists": false}}
	err := r.db.Collection(userCollection).FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}
