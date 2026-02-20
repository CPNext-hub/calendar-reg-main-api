package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ----- Mock UserRepository -----

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) GetPaginated(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

// ----- Tests -----

func TestSeedSuperAdmin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "admin").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	uc.SeedSuperAdmin(context.Background(), "admin", "pass")
	repo.AssertExpectations(t)
}

func TestSeedSuperAdmin_AlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "admin").Return(&entity.User{}, nil)

	uc.SeedSuperAdmin(context.Background(), "admin", "pass")
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSeedSuperAdmin_FindError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "admin").Return(nil, errors.New("db error"))

	uc.SeedSuperAdmin(context.Background(), "admin", "pass")
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSeedSuperAdmin_CreateError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "admin").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(errors.New("create error"))

	uc.SeedSuperAdmin(context.Background(), "admin", "pass")
	repo.AssertExpectations(t)
}

func TestSeedSuperAdmin_HashError(t *testing.T) {
	orig := hashPassword
	hashPassword = func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("hash error")
	}
	defer func() { hashPassword = orig }()

	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "admin").Return(nil, nil)

	uc.SeedSuperAdmin(context.Background(), "admin", "pass")
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRegister_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == "user1" && u.Role == constants.RoleStudent
	})).Return(nil)

	user, err := uc.Register(context.Background(), "user1", "pass", constants.RoleStudent, nil)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestRegister_InvalidRole(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	_, err := uc.Register(context.Background(), "user1", "pass", "invalid_role", nil)
	assert.EqualError(t, err, "invalid role")
}

func TestRegister_SuperAdminSelfRegister(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	_, err := uc.Register(context.Background(), "super", "pass", "superadmin", nil)
	assert.EqualError(t, err, "superadmin cannot be created via registration")
}

func TestRegister_CreateAdmin_Unauthorized(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	caller := "user"
	_, err := uc.Register(context.Background(), "newadmin", "pass", "admin", &caller)
	assert.EqualError(t, err, "only superadmin or admin can create admin users")

	// No caller
	_, err = uc.Register(context.Background(), "newadmin", "pass", "admin", nil)
	assert.EqualError(t, err, "only superadmin or admin can create admin users")
}

func TestRegister_CreateAdmin_Authorized(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "newadmin").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	caller := "superadmin"
	user, err := uc.Register(context.Background(), "newadmin", "pass", "admin", &caller)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(&entity.User{}, nil)

	_, err := uc.Register(context.Background(), "user1", "pass", constants.RoleStudent, nil)
	assert.EqualError(t, err, "username already exists")
}

func TestRegister_FindError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, errors.New("db error"))

	_, err := uc.Register(context.Background(), "user1", "pass", constants.RoleStudent, nil)
	assert.EqualError(t, err, "db error")
}

func TestRegister_CreateError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Register(context.Background(), "user1", "pass", constants.RoleStudent, nil)
	assert.EqualError(t, err, "db error")
}

func TestRegister_HashError(t *testing.T) {
	orig := hashPassword
	hashPassword = func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("hash error")
	}
	defer func() { hashPassword = orig }()

	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, nil)

	_, err := uc.Register(context.Background(), "user1", "pass", constants.RoleStudent, nil)
	assert.EqualError(t, err, "hash error")
}

func TestLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user := &entity.User{BaseEntity: entity.BaseEntity{ID: "u1"}, Username: "user1", Password: string(hashed), Role: "user"}

	repo.On("FindByUsername", mock.Anything, "user1").Return(user, nil)

	token, err := uc.Login(context.Background(), "user1", "pass")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, nil)

	_, err := uc.Login(context.Background(), "user1", "pass")
	assert.EqualError(t, err, "invalid credentials")
}

func TestLogin_FindError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("FindByUsername", mock.Anything, "user1").Return(nil, errors.New("db error"))

	_, err := uc.Login(context.Background(), "user1", "pass")
	assert.EqualError(t, err, "db error")
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user := &entity.User{BaseEntity: entity.BaseEntity{ID: "u1"}, Username: "user1", Password: string(hashed), Role: "user"}

	repo.On("FindByUsername", mock.Anything, "user1").Return(user, nil)

	_, err := uc.Login(context.Background(), "user1", "wrongpass")
	assert.EqualError(t, err, "invalid credentials")
}

func TestLogin_SignTokenError(t *testing.T) {
	orig := signToken
	signToken = func(token *jwt.Token, secret []byte) (string, error) {
		return "", errors.New("sign error")
	}
	defer func() { signToken = orig }()

	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user := &entity.User{BaseEntity: entity.BaseEntity{ID: "u1"}, Username: "user1", Password: string(hashed), Role: "user"}

	repo.On("FindByUsername", mock.Anything, "user1").Return(user, nil)

	_, err := uc.Login(context.Background(), "user1", "pass")
	assert.EqualError(t, err, "sign error")
}

func TestGetUsersPaginated_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	users := []*entity.User{{Username: "u1"}, {Username: "u2"}}
	repo.On("GetPaginated", mock.Anything, 1, 10).Return(users, int64(2), nil)

	res, err := uc.GetUsersPaginated(context.Background(), pagination.PaginationQuery{Page: 1, Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, int64(2), res.Total)
}

func TestGetUsersPaginated_Error(t *testing.T) {
	repo := new(mockUserRepo)
	uc := NewAuthUsecase(repo, "secret")

	repo.On("GetPaginated", mock.Anything, 1, 10).Return(nil, int64(0), errors.New("db error"))

	_, err := uc.GetUsersPaginated(context.Background(), pagination.PaginationQuery{Page: 1, Limit: 10})
	assert.EqualError(t, err, "db error")
}
