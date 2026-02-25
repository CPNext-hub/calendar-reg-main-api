package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Package-level function vars for testability.
var hashPassword = func(password []byte, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

var signToken = func(token *jwt.Token, secret []byte) (string, error) {
	return token.SignedString(secret)
}

// AuthUsecase defines the business logic for authentication.
type AuthUsecase interface {
	Register(ctx context.Context, username, password string, role string, callerRole *string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	SeedSuperAdmin(ctx context.Context, username, password string)
	GetUsersPaginated(ctx context.Context, pq pagination.PaginationQuery) (*pagination.PaginatedResult[*entity.User], error)
	DeleteUser(ctx context.Context, id string, callerRole string, callerID string) error
	UpdateUserRole(ctx context.Context, id string, newRole string, callerRole string, callerID string) (*entity.User, error)
}

type authUsecase struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

// NewAuthUsecase creates a new instance of AuthUsecase.
func NewAuthUsecase(repo repository.UserRepository, jwtSecret string) AuthUsecase {
	return &authUsecase{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

// SeedSuperAdmin creates the default superadmin user if one doesn't already exist.
func (u *authUsecase) SeedSuperAdmin(ctx context.Context, username, password string) {
	existing, err := u.repo.FindByUsername(ctx, username)
	if err != nil {
		log.Printf("Warning: failed to check for existing superadmin: %v", err)
		return
	}
	if existing != nil {
		log.Printf("Superadmin user '%s' already exists, skipping seed", username)
		return
	}

	hashed, err := hashPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning: failed to hash superadmin password: %v", err)
		return
	}

	user := &entity.User{
		Username: username,
		Password: string(hashed),
		Role:     constants.RoleSuperAdmin,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		log.Printf("Warning: failed to seed superadmin: %v", err)
		return
	}

	log.Printf("Superadmin user '%s' created successfully", username)
}

func (u *authUsecase) Register(ctx context.Context, username, password string, role string, callerRole *string) (*entity.User, error) {
	// Validate role
	if !constants.ValidRoles[role] {
		return nil, errors.New("invalid role")
	}

	// Prevent self-registration of superadmin
	if role == constants.RoleSuperAdmin {
		return nil, errors.New("superadmin cannot be created via registration")
	}

	// Creating admin requires caller to be superadmin or admin
	if role == constants.RoleAdmin {
		if callerRole == nil || !constants.PrivilegedRoles[*callerRole] {
			return nil, errors.New("only superadmin or admin can create admin users")
		}
	}

	// Check if user already exists
	existing, err := u.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashed, err := hashPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username: username,
		Password: string(hashed),
		Role:     role,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := u.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := signToken(token, u.jwtSecret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (u *authUsecase) GetUsersPaginated(ctx context.Context, pq pagination.PaginationQuery) (*pagination.PaginatedResult[*entity.User], error) {
	items, total, err := u.repo.GetPaginated(ctx, pq.Page, pq.Limit)
	if err != nil {
		return nil, err
	}
	result := pagination.NewResult(items, pq.Page, pq.Limit, total)
	return &result, nil
}

// DeleteUser soft-deletes a user. Callers cannot delete themselves or superadmins (unless caller is superadmin).
func (u *authUsecase) DeleteUser(ctx context.Context, id string, callerRole string, callerID string) error {
	if id == callerID {
		return errors.New("cannot delete your own account")
	}

	target, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("user not found")
	}

	// Only superadmin can delete superadmin accounts
	if target.Role == constants.RoleSuperAdmin {
		return errors.New("superadmin accounts cannot be deleted")
	}

	// Admin can only delete student accounts
	if callerRole == constants.RoleAdmin && target.Role == constants.RoleAdmin {
		return errors.New("admin cannot delete another admin account")
	}

	return u.repo.DeleteUser(ctx, id)
}

// UpdateUserRole changes the role of a user. Callers cannot change their own role.
func (u *authUsecase) UpdateUserRole(ctx context.Context, id string, newRole string, callerRole string, callerID string) (*entity.User, error) {
	if id == callerID {
		return nil, errors.New("cannot change your own role")
	}

	// Validate new role
	if !constants.ValidRoles[newRole] {
		return nil, errors.New("invalid role")
	}

	// Only superadmin can assign/change superadmin role
	if newRole == constants.RoleSuperAdmin {
		return nil, errors.New("cannot assign superadmin role")
	}

	target, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, errors.New("user not found")
	}

	// Superadmin accounts cannot have their role changed
	if target.Role == constants.RoleSuperAdmin {
		return nil, errors.New("superadmin role cannot be changed")
	}

	// Admin can only manage students
	if callerRole == constants.RoleAdmin && target.Role == constants.RoleAdmin {
		return nil, errors.New("admin cannot change another admin's role")
	}

	if err := u.repo.UpdateUserRole(ctx, id, newRole); err != nil {
		return nil, err
	}

	target.Role = newRole
	return target, nil
}
