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

const cronJobCollection = "cronjobs"

// cronJobModel is the MongoDB-specific representation of a CronJob.
type cronJobModel struct {
	BaseModel   `bson:",inline"`
	Name        string   `bson:"name"`
	CourseCodes []string `bson:"course_codes"`
	Acadyear    int      `bson:"acadyear"`
	Semester    int      `bson:"semester"`
	CronExpr    string   `bson:"cron_expr"`
	Enabled     bool     `bson:"enabled"`
}

// toEntity converts a MongoDB model to a domain entity.
func (m *cronJobModel) toEntity() *entity.CronJob {
	var id string
	if m.ID != nil {
		id = m.ID.Hex()
	}

	return &entity.CronJob{
		BaseEntity: entity.BaseEntity{
			ID:        id,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
		Name:        m.Name,
		CourseCodes: m.CourseCodes,
		Acadyear:    m.Acadyear,
		Semester:    m.Semester,
		CronExpr:    m.CronExpr,
		Enabled:     m.Enabled,
	}
}

// toCronJobModel converts a domain entity to a MongoDB model.
func toCronJobModel(e *entity.CronJob) *cronJobModel {
	m := &cronJobModel{
		Name:        e.Name,
		CourseCodes: e.CourseCodes,
		Acadyear:    e.Acadyear,
		Semester:    e.Semester,
		CronExpr:    e.CronExpr,
		Enabled:     e.Enabled,
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

type cronJobRepository struct {
	db *mongo.Database
}

// NewCronJobRepository creates a new instance of CronJobRepository.
func NewCronJobRepository(db *mongo.Database) repository.CronJobRepository {
	return &cronJobRepository{db: db}
}

func (r *cronJobRepository) Create(ctx context.Context, job *entity.CronJob) error {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	model := toCronJobModel(job)
	result, err := r.db.Collection(cronJobCollection).InsertOne(ctx, model)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		job.ID = oid.Hex()
	}
	return nil
}

func (r *cronJobRepository) GetAll(ctx context.Context) ([]*entity.CronJob, error) {
	cursor, err := r.db.Collection(cronJobCollection).Find(ctx, notDeleted)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*cronJobModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	jobs := make([]*entity.CronJob, len(models))
	for i, m := range models {
		jobs[i] = m.toEntity()
	}
	return jobs, nil
}

func (r *cronJobRepository) GetByID(ctx context.Context, id string) (*entity.CronJob, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	filter := bson.M{
		"_id":        oid,
		"deleted_at": bson.M{"$exists": false},
	}

	var model cronJobModel
	err = r.db.Collection(cronJobCollection).FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

func (r *cronJobRepository) Update(ctx context.Context, job *entity.CronJob) error {
	job.UpdatedAt = time.Now()

	oid, err := bson.ObjectIDFromHex(job.ID)
	if err != nil {
		return errors.New("invalid id format")
	}

	filter := bson.M{
		"_id":        oid,
		"deleted_at": bson.M{"$exists": false},
	}
	update := bson.M{
		"$set": bson.M{
			"name":         job.Name,
			"course_codes": job.CourseCodes,
			"acadyear":     job.Acadyear,
			"semester":     job.Semester,
			"cron_expr":    job.CronExpr,
			"enabled":      job.Enabled,
			"updated_at":   job.UpdatedAt,
		},
	}

	result, err := r.db.Collection(cronJobCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("cron job not found")
	}
	return nil
}

func (r *cronJobRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	now := time.Now()
	filter := bson.M{
		"_id":        oid,
		"deleted_at": bson.M{"$exists": false},
	}
	update := bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}}

	result, err := r.db.Collection(cronJobCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("cron job not found")
	}
	return nil
}

func (r *cronJobRepository) GetEnabled(ctx context.Context) ([]*entity.CronJob, error) {
	filter := bson.M{
		"enabled":    true,
		"deleted_at": bson.M{"$exists": false},
	}

	cursor, err := r.db.Collection(cronJobCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*cronJobModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	jobs := make([]*entity.CronJob, len(models))
	for i, m := range models {
		jobs[i] = m.toEntity()
	}
	return jobs, nil
}
