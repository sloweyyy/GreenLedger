package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email       string    `gorm:"uniqueIndex;not null" json:"email"`
	Username    string    `gorm:"uniqueIndex;not null" json:"username"`
	FirstName   string    `gorm:"not null" json:"first_name"`
	LastName    string    `gorm:"not null" json:"last_name"`
	Password    string    `gorm:"not null" json:"-"` // Never include in JSON
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	IsVerified  bool      `gorm:"default:false" json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	Roles    []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Sessions []Session `gorm:"foreignKey:UserID" json:"-"`
	Profile  *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Resource    string    `gorm:"not null" json:"resource"`
	Action      string    `gorm:"not null" json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationship
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// UserProfile represents extended user profile information
type UserProfile struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	PhoneNumber string    `json:"phone_number"`
	Preferences map[string]interface{} `gorm:"type:jsonb" json:"preferences"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationship
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
	
	// Relationship
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
	
	// Relationship
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// GetRoleNames returns a slice of role names for the user
func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roleNames[i] = role.Name
	}
	return roleNames
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(resource, action string) bool {
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if permission.Resource == resource && permission.Action == action {
				return true
			}
		}
	}
	return false
}

// IsSessionValid checks if a session is still valid
func (s *Session) IsSessionValid() bool {
	return s.IsActive && time.Now().Before(s.ExpiresAt)
}

// Table names
func (User) TableName() string                      { return "users" }
func (Role) TableName() string                      { return "roles" }
func (Permission) TableName() string                { return "permissions" }
func (Session) TableName() string                   { return "sessions" }
func (UserProfile) TableName() string               { return "user_profiles" }
func (PasswordResetToken) TableName() string        { return "password_reset_tokens" }
func (EmailVerificationToken) TableName() string    { return "email_verification_tokens" }

// Constants for default roles
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleModerator = "moderator"
)

// Constants for permissions
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionAdmin  = "admin"
)

// Constants for resources
const (
	ResourceUser        = "user"
	ResourceCalculation = "calculation"
	ResourceWallet      = "wallet"
	ResourceReport      = "report"
	ResourceCertificate = "certificate"
)
