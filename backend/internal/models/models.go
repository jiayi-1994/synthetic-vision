package models

import "time"

// User is both the persistence model and the JSON DTO returned to clients.
// PasswordHash and IsDemo are never serialized.
type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	PublicID     string    `gorm:"uniqueIndex;size:16" json:"public_id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:160" json:"email"`
	PasswordHash string    `gorm:"size:120" json:"-"`
	Role         string    `gorm:"size:16;default:user" json:"role"`
	Plan         string    `gorm:"size:16;default:free" json:"plan"`
	Credits      int       `gorm:"default:0" json:"credits"`
	AvatarSeed   string    `gorm:"size:64" json:"avatar_seed"`
	IsDemo       bool      `gorm:"default:false" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Generation is both the persistence model and the JSON DTO for an image job.
type Generation struct {
	ID             string     `gorm:"primaryKey;size:36" json:"id"`
	UserID         string     `gorm:"index;size:36" json:"-"`
	Prompt         string     `gorm:"type:text" json:"prompt"`
	NegativePrompt string     `gorm:"type:text" json:"negative_prompt"`
	Resolution     string     `gorm:"size:8" json:"resolution"`
	AspectRatio    string     `gorm:"size:8" json:"aspect_ratio"`
	Style          string     `gorm:"size:32" json:"style"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	Seed           int64      `json:"seed"`
	Status         string     `gorm:"size:16;index" json:"status"`
	Cost           int        `json:"cost"`
	ImageURL       string     `gorm:"size:255" json:"image_url"` // e.g. "/images/<id>.png"
	Error          string     `gorm:"type:text" json:"error"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// CreditTransaction is an immutable ledger entry. Amount is positive for a
// credit and negative for a debit. Reason is one of:
// generation | admin_injection | signup_bonus | refund.
type CreditTransaction struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"index;size:36" json:"user_id"`
	Amount    int       `json:"amount"`                // + credit, - debit
	Reason    string    `gorm:"size:32" json:"reason"` // generation|admin_injection|signup_bonus|refund
	AdminID   string    `gorm:"size:36" json:"admin_id"`
	CreatedAt time.Time `json:"created_at"`
}
