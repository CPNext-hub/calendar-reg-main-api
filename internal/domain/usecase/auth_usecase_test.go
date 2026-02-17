package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ----- mock UserRepository -----

type mockUserRepo struct {
	users     map[string]*entity.User
	createErr error
	findErr   error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*entity.User)}
}

func (m *mockUserRepo) Create(_ context.Context, u *entity.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[u.Username] = u
	return nil
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*entity.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	u, ok := m.users[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) GetPaginated(_ context.Context, page, limit int) ([]*entity.User, int64, error) {
	var users []*entity.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, int64(len(users)), nil
}

const testJWTSecret = "test-secret-key"

// ----- Register tests -----

func TestRegister_StudentSuccess(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	user, err := uc.Register(context.Background(), "john", "pass123", "student", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "john" {
		t.Errorf("expected username='john', got %q", user.Username)
	}
	if user.Role != "student" {
		t.Errorf("expected role='student', got %q", user.Role)
	}
	// Password should be hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("pass123")); err != nil {
		t.Error("password was not properly hashed")
	}
}

func TestRegister_AdminByPrivilegedCaller(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	callerRole := "superadmin"
	user, err := uc.Register(context.Background(), "admin1", "pass", "admin", &callerRole)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != "admin" {
		t.Errorf("expected role='admin', got %q", user.Role)
	}
}

func TestRegister_AdminByNonPrivileged(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	callerRole := "student"
	_, err := uc.Register(context.Background(), "admin1", "pass", "admin", &callerRole)
	if err == nil {
		t.Fatal("expected error for non-privileged caller creating admin")
	}
	if !strings.Contains(err.Error(), "only superadmin or admin") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_AdminWithoutCaller(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "admin1", "pass", "admin", nil)
	if err == nil {
		t.Fatal("expected error for nil caller creating admin")
	}
}

func TestRegister_SuperAdminForbidden(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "sa", "pass", "superadmin", nil)
	if err == nil {
		t.Fatal("expected error for superadmin registration")
	}
	if !strings.Contains(err.Error(), "superadmin cannot be created") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "user1", "pass", "unknown_role", nil)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if err.Error() != "invalid role" {
		t.Errorf("expected 'invalid role', got %q", err.Error())
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	repo := newMockUserRepo()
	repo.users["john"] = &entity.User{Username: "john"}
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "john", "pass", "student", nil)
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	if !strings.Contains(err.Error(), "username already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_RepoFindError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findErr = errors.New("db error")
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "john", "pass", "student", nil)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestRegister_RepoCreateError(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("insert failed")
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Register(context.Background(), "john", "pass", "student", nil)
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %v", err)
	}
}

// ----- Login tests -----

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	repo.users["john"] = &entity.User{
		BaseEntity: entity.BaseEntity{ID: "user123"},
		Username:   "john",
		Password:   string(hashed),
		Role:       "student",
	}
	uc := NewAuthUsecase(repo, testJWTSecret)

	tokenStr, err := uc.Login(context.Background(), "john", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	// Verify token claims
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		t.Fatal("invalid token")
	}
	if claims["username"] != "john" {
		t.Errorf("expected username='john' in claims, got %v", claims["username"])
	}
	if claims["role"] != "student" {
		t.Errorf("expected role='student' in claims, got %v", claims["role"])
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Login(context.Background(), "nobody", "pass")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %q", err.Error())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.users["john"] = &entity.User{Username: "john", Password: string(hashed)}
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Login(context.Background(), "john", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %q", err.Error())
	}
}

func TestLogin_RepoError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findErr = errors.New("db error")
	uc := NewAuthUsecase(repo, testJWTSecret)

	_, err := uc.Login(context.Background(), "john", "pass")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

// ----- SeedSuperAdmin tests -----

func TestSeedSuperAdmin_CreatesWhenNotExists(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewAuthUsecase(repo, testJWTSecret)

	uc.SeedSuperAdmin(context.Background(), "admin", "pass123")

	user, ok := repo.users["admin"]
	if !ok {
		t.Fatal("expected superadmin to be created")
	}
	if user.Role != "superadmin" {
		t.Errorf("expected role='superadmin', got %q", user.Role)
	}
	// Password should be hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("pass123")); err != nil {
		t.Error("password was not properly hashed")
	}
}

func TestSeedSuperAdmin_SkipsWhenExists(t *testing.T) {
	repo := newMockUserRepo()
	existing := &entity.User{Username: "admin", Role: "superadmin"}
	repo.users["admin"] = existing
	uc := NewAuthUsecase(repo, testJWTSecret)

	uc.SeedSuperAdmin(context.Background(), "admin", "newpass")

	// Should still be the old user, not recreated
	if repo.users["admin"] != existing {
		t.Error("expected existing superadmin to remain unchanged")
	}
}
