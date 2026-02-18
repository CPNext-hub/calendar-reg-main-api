package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
)

// ----- mock CourseRepository -----

type mockCourseRepo struct {
	courses    map[string]*entity.Course // key = "code:year:semester"
	createErr  error
	getByErr   error
	getAllErr  error
	pagErr     error
	updateErr  error
	deleteErr  error
	allCourses []*entity.Course
}

func newMockCourseRepo() *mockCourseRepo {
	return &mockCourseRepo{courses: make(map[string]*entity.Course)}
}

func mockKey(code string, year, semester int) string {
	return fmt.Sprintf("%s:%d:%d", code, year, semester)
}

func (m *mockCourseRepo) Create(_ context.Context, c *entity.Course) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.courses[c.Key()] = c
	return nil
}

func (m *mockCourseRepo) GetAll(_ context.Context) ([]*entity.Course, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.allCourses, nil
}

func (m *mockCourseRepo) GetPaginated(_ context.Context, page, limit int, includeSections bool) ([]*entity.Course, int64, error) {
	if m.pagErr != nil {
		return nil, 0, m.pagErr
	}
	return m.allCourses, int64(len(m.allCourses)), nil
}

func (m *mockCourseRepo) GetByKey(_ context.Context, code string, year, semester int) (*entity.Course, error) {
	if m.getByErr != nil {
		return nil, m.getByErr
	}
	c, ok := m.courses[mockKey(code, year, semester)]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockCourseRepo) SoftDelete(_ context.Context, code string, year, semester int) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.courses, mockKey(code, year, semester))
	return nil
}

func (m *mockCourseRepo) Update(_ context.Context, c *entity.Course) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.courses[c.Key()] = c
	return nil
}

// ----- CreateCourse tests -----

func TestCreateCourse_Success(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	course := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	err := uc.CreateCourse(context.Background(), course)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, ok := repo.courses[course.Key()]; !ok {
		t.Error("expected course to be stored in repo")
	}
}

func TestCreateCourse_AlreadyExists(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil {
		t.Fatal("expected error for duplicate course")
	}
}

func TestCreateCourse_RepoGetByKeyError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getByErr = errors.New("db error")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestCreateCourse_RepoCreateError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.createErr = errors.New("insert failed")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %v", err)
	}
}

// ----- GetAllCourses tests -----

func TestGetAllCourses_Success(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}, {Code: "CS102"}}
	uc := NewCourseUsecase(repo, nil, nil)

	courses, err := uc.GetAllCourses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetAllCourses_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getAllErr = errors.New("find failed")
	uc := NewCourseUsecase(repo, nil, nil)

	_, err := uc.GetAllCourses(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- GetCoursesPaginated tests -----

func TestGetCoursesPaginated_Success(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}, {Code: "CS102"}, {Code: "CS103"}}
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 10}
	result, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected total=3, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
	if result.Page != 1 {
		t.Errorf("expected page=1, got %d", result.Page)
	}
}

func TestGetCoursesPaginated_LimitZero(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}}
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 0}
	result, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPages != 1 {
		t.Errorf("expected totalPages=1 for limit=0, got %d", result.TotalPages)
	}
}

func TestGetCoursesPaginated_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.pagErr = errors.New("paginate failed")
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 10}
	_, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- GetCourseByCode tests -----

func TestGetCourseByCode_Found(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1, NameEN: "Intro CS"}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	course, err := uc.GetCourseByCode(context.Background(), "CS101", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course == nil || course.NameEN != "Intro CS" {
		t.Error("expected to find course with correct name")
	}
}

func TestGetCourseByCode_NotFound(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	course, err := uc.GetCourseByCode(context.Background(), "NOPE", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course != nil {
		t.Error("expected nil for non-existent course")
	}
}

func TestGetCourseByCode_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getByErr = errors.New("db error")
	uc := NewCourseUsecase(repo, nil, nil)

	_, err := uc.GetCourseByCode(context.Background(), "CS101", 2568, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- DeleteCourse tests -----

func TestDeleteCourse_Success(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "CS101", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.courses[c.Key()]; ok {
		t.Error("expected course to be deleted from repo")
	}
}

func TestDeleteCourse_NotFound(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "NOPE", 2568, 1)
	if err == nil {
		t.Fatal("expected error for non-existent course")
	}
	if err.Error() != "course not found" {
		t.Errorf("expected 'course not found', got %q", err.Error())
	}
}

func TestDeleteCourse_RepoError(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	repo.deleteErr = errors.New("delete failed")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "CS101", 2568, 1)
	if err == nil || err.Error() != "delete failed" {
		t.Errorf("expected 'delete failed', got %v", err)
	}
}
