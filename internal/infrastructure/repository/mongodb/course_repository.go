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

const courseCollection = "courses"

// courseModel is the MongoDB-specific representation of a course (bson tags live here).
type courseModel struct {
	baseModel `bson:",inline"`
	Code      string `bson:"code"`
	Name      string `bson:"name"`
	Credits   string `bson:"credits"`
}

// toEntity converts a MongoDB model to a domain entity.
func (m *courseModel) toEntity() *entity.Course {
	return &entity.Course{
		BaseEntity: entity.BaseEntity{
			ID:        m.ID.Hex(),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
		Code:    m.Code,
		Name:    m.Name,
		Credits: m.Credits,
	}
}

// toCourseModel converts a domain entity to a MongoDB model.
func toCourseModel(e *entity.Course) *courseModel {
	m := &courseModel{
		Code:    e.Code,
		Name:    e.Name,
		Credits: e.Credits,
	}
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = e.DeletedAt
	if e.ID != "" {
		oid, err := bson.ObjectIDFromHex(e.ID)
		if err == nil {
			m.ID = oid
		}
	}
	return m
}

type courseRepository struct {
	db *mongo.Database
}

// NewCourseRepository creates a new instance of CourseRepository.
func NewCourseRepository(db *mongo.Database) repository.CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(ctx context.Context, course *entity.Course) error {
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	model := toCourseModel(course)
	_, err := r.db.Collection(courseCollection).InsertOne(ctx, model)
	return err
}

// notDeleted is the filter to exclude soft-deleted documents.
var notDeleted = bson.M{"deleted_at": bson.M{"$exists": false}}

func (r *courseRepository) GetAll(ctx context.Context) ([]*entity.Course, error) {
	cursor, err := r.db.Collection(courseCollection).Find(ctx, notDeleted)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*courseModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	courses := make([]*entity.Course, len(models))
	for i, m := range models {
		courses[i] = m.toEntity()
	}
	return courses, nil
}

func (r *courseRepository) GetByCode(ctx context.Context, code string) (*entity.Course, error) {
	var model courseModel
	filter := bson.M{"code": code, "deleted_at": bson.M{"$exists": false}}
	err := r.db.Collection(courseCollection).FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

func (r *courseRepository) SoftDelete(ctx context.Context, code string) error {
	now := time.Now()
	filter := bson.M{"code": code, "deleted_at": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}}

	result, err := r.db.Collection(courseCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
