package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTeacher UserRole = "teacher"
	RoleStudent UserRole = "student"
)

type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Username  string     `json:"username" db:"username"`
	Password  string     `json:"-" db:"password"`
	FullName  string     `json:"full_name" db:"full_name"`
	AvatarURL *string    `json:"avatar_url" db:"avatar_url"`
	Role      UserRole   `json:"role" db:"role"`
	SchoolID  *uuid.UUID `json:"school_id" db:"school_id"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url"`
	Role      UserRole  `json:"role"`
}

func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
	}
}

type RegisterRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Password string   `json:"password" validate:"required,min=8"`
	FullName string   `json:"full_name" validate:"required"`
	Role     UserRole `json:"role" validate:"required,oneof=teacher student"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         UserPublic `json:"user"`
}

type Class struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TeacherID uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Name      string    `json:"name" db:"name"`
	Grade     string    `json:"grade" db:"grade"`
	Subject   string    `json:"subject" db:"subject"`
	ClassCode string    `json:"class_code" db:"class_code"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	StudentCount int `json:"student_count,omitempty" db:"student_count"`
}
