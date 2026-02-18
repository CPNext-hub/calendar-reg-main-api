package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const courseCollection = "courses"

// courseModel is the MongoDB-specific representation of a course (bson tags live here).
type courseModel struct {
	BaseModel    `bson:",inline"`
	Code         string         `bson:"code"`
	NameEN       string         `bson:"name_en"`
	NameTH       string         `bson:"name_th"`
	Faculty      string         `bson:"faculty"`
	Department   string         `bson:"department,omitempty"`
	Credits      string         `bson:"credits"`
	Prerequisite string         `bson:"prerequisite,omitempty"`
	Semester     int            `bson:"semester"`
	Year         int            `bson:"year"`
	Sections     []sectionModel `bson:"sections"`
}

type sectionModel struct {
	Number       string          `bson:"number"`
	Schedules    []scheduleModel `bson:"schedules"`
	Seats        int             `bson:"seats"`
	Instructor   string          `bson:"instructor"`
	ExamStart    time.Time       `bson:"exam_start,omitempty"`
	ExamEnd      time.Time       `bson:"exam_end,omitempty"`
	MidtermStart time.Time       `bson:"midterm_start,omitempty"`
	MidtermEnd   time.Time       `bson:"midterm_end,omitempty"`
	Note         string          `bson:"note,omitempty"`
	ReservedFor  []string        `bson:"reserved_for,omitempty"`
	Campus       string          `bson:"campus,omitempty"`
	Program      string          `bson:"program,omitempty"`
}

type scheduleModel struct {
	Day       string    `bson:"day"`
	StartTime time.Time `bson:"start_time"`
	EndTime   time.Time `bson:"end_time"`
	Room      string    `bson:"room"`
	Type      string    `bson:"type"`
}

// compositeFilter builds the composite key filter for lookups.
func compositeFilter(code string, year, semester int) bson.M {
	return bson.M{
		"code":       code,
		"year":       year,
		"semester":   semester,
		"deleted_at": bson.M{"$exists": false},
	}
}

// toEntity converts a MongoDB model to a domain entity.
func (m *courseModel) toEntity() *entity.Course {
	sections := make([]entity.Section, len(m.Sections))
	for i, s := range m.Sections {
		schedules := make([]entity.Schedule, len(s.Schedules))
		for j, sc := range s.Schedules {
			schedules[j] = entity.Schedule{
				Day:       sc.Day,
				StartTime: sc.StartTime,
				EndTime:   sc.EndTime,
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}
		sections[i] = entity.Section{
			Number:       s.Number,
			Schedules:    schedules,
			Seats:        s.Seats,
			Instructor:   s.Instructor,
			ExamStart:    s.ExamStart,
			ExamEnd:      s.ExamEnd,
			MidtermStart: s.MidtermStart,
			MidtermEnd:   s.MidtermEnd,
			Note:         s.Note,
			ReservedFor:  s.ReservedFor,
			Campus:       s.Campus,
			Program:      s.Program,
		}
	}

	var id string
	if m.ID != nil {
		id = m.ID.Hex()
	}

	return &entity.Course{
		BaseEntity: entity.BaseEntity{
			ID:        id,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
		Code:         m.Code,
		NameEN:       m.NameEN,
		NameTH:       m.NameTH,
		Faculty:      m.Faculty,
		Department:   m.Department,
		Credits:      m.Credits,
		Prerequisite: m.Prerequisite,
		Semester:     m.Semester,
		Year:         m.Year,
		Sections:     sections,
	}
}

// toCourseModel converts a domain entity to a MongoDB model.
func toCourseModel(e *entity.Course) *courseModel {
	sections := make([]sectionModel, len(e.Sections))
	for i, s := range e.Sections {
		schedules := make([]scheduleModel, len(s.Schedules))
		for j, sc := range s.Schedules {
			schedules[j] = scheduleModel{
				Day:       sc.Day,
				StartTime: sc.StartTime,
				EndTime:   sc.EndTime,
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}
		sections[i] = sectionModel{
			Number:       s.Number,
			Schedules:    schedules,
			Seats:        s.Seats,
			Instructor:   s.Instructor,
			ExamStart:    s.ExamStart,
			ExamEnd:      s.ExamEnd,
			MidtermStart: s.MidtermStart,
			MidtermEnd:   s.MidtermEnd,
			Note:         s.Note,
			ReservedFor:  s.ReservedFor,
			Campus:       s.Campus,
			Program:      s.Program,
		}
	}

	m := &courseModel{
		Code:         e.Code,
		NameEN:       e.NameEN,
		NameTH:       e.NameTH,
		Faculty:      e.Faculty,
		Department:   e.Department,
		Credits:      e.Credits,
		Prerequisite: e.Prerequisite,
		Semester:     e.Semester,
		Year:         e.Year,
		Sections:     sections,
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
	result, err := r.db.Collection(courseCollection).InsertOne(ctx, model)
	if err != nil {
		return err
	}

	// Write back the generated ID to the entity.
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		course.ID = oid.Hex()
	}
	return nil
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

func (r *courseRepository) GetPaginated(ctx context.Context, page, limit int, includeSections bool) ([]*entity.Course, int64, error) {
	col := r.db.Collection(courseCollection)

	// Count total matching documents.
	total, err := col.CountDocuments(ctx, notDeleted)
	if err != nil {
		return nil, 0, err
	}

	// Build find options.
	opts := options.Find()
	if !includeSections {
		opts.SetProjection(bson.M{"sections": 0})
	}
	if limit > 0 {
		skip := int64((page - 1) * limit)
		opts.SetSkip(skip)
		opts.SetLimit(int64(limit))
	}
	// limit == 0 → no skip/limit → return all

	cursor, err := col.Find(ctx, notDeleted, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var models []*courseModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, 0, err
	}

	courses := make([]*entity.Course, len(models))
	for i, m := range models {
		courses[i] = m.toEntity()
	}
	return courses, total, nil
}

func (r *courseRepository) GetByKey(ctx context.Context, code string, year, semester int) (*entity.Course, error) {
	var model courseModel
	filter := compositeFilter(code, year, semester)
	err := r.db.Collection(courseCollection).FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

func (r *courseRepository) Update(ctx context.Context, course *entity.Course) error {
	course.UpdatedAt = time.Now()
	model := toCourseModel(course)

	filter := compositeFilter(course.Code, course.Year, course.Semester)
	update := bson.M{
		"$set": bson.M{
			"name_en":      model.NameEN,
			"name_th":      model.NameTH,
			"faculty":      model.Faculty,
			"department":   model.Department,
			"credits":      model.Credits,
			"prerequisite": model.Prerequisite,
			"semester":     model.Semester,
			"year":         model.Year,
			"sections":     model.Sections,
			"updated_at":   model.UpdatedAt,
		},
	}

	result, err := r.db.Collection(courseCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("course not found")
	}
	return nil
}

func (r *courseRepository) SoftDelete(ctx context.Context, code string, year, semester int) error {
	now := time.Now()
	filter := compositeFilter(code, year, semester)
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
