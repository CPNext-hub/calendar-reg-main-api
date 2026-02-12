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
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Code      string        `bson:"code"`
	Name      string        `bson:"name"`
	Credits   string        `bson:"credits"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

// toEntity converts a MongoDB model to a domain entity.
func (m *courseModel) toEntity() *entity.Course {
	return &entity.Course{
		ID:        m.ID.Hex(),
		Code:      m.Code,
		Name:      m.Name,
		Credits:   m.Credits,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// toCourseModel converts a domain entity to a MongoDB model.
func toCourseModel(e *entity.Course) *courseModel {
	m := &courseModel{
		Code:      e.Code,
		Name:      e.Name,
		Credits:   e.Credits,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
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

func (r *courseRepository) GetAll(ctx context.Context) ([]*entity.Course, error) {
	cursor, err := r.db.Collection(courseCollection).Find(ctx, bson.M{})
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
	err := r.db.Collection(courseCollection).FindOne(ctx, bson.M{"code": code}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

func (r *courseRepository) Delete(ctx context.Context, code string) error {
	result, err := r.db.Collection(courseCollection).DeleteOne(ctx, bson.M{"code": code})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
