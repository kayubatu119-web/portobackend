package model

import (
	"time"

	"github.com/google/uuid"
)

// ============================
// MAIN EXPERIENCE MODEL
// ============================

type Experience struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title        string    `json:"title" gorm:"type:varchar(200);not null"`
	Company      string    `json:"company" gorm:"type:varchar(150);not null"`
	Location     string    `json:"location" gorm:"type:varchar(200);not null"`
	StartYears   string    `json:"start_year" gorm:"column:start_year;type:varchar(20);not null"`
	EndYears     string    `json:"end_year" gorm:"column:end_year;type:varchar(20);not null"`
	CurrentJob   bool      `json:"current_job" gorm:"column:current_job;type:boolean;default:false"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

type ExperienceWithRelations struct {
	Experience
	Responsibilities []ExperienceResponsibility `json:"responsibilities"`
	Skills           []ExperienceSkill          `json:"skills"`
}

func (Experience) TableName() string {
	return "portfolio_experiences"
}

// ============================
// RESPONSIBILITY MODEL
// ============================

type ExperienceResponsibility struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"` // Biarkan database generate
	ExperienceID uuid.UUID `json:"experience_id" gorm:"type:uuid;not null"`
	Description  string    `json:"description" gorm:"type:text;not null"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (ExperienceResponsibility) TableName() string {
	return "experience_responsibilities"
}

// ============================
// SKILL MODEL - PERBAIKAN: Hapus CreatedAt
// ============================

type ExperienceSkill struct {
	ExperienceID uuid.UUID `json:"experience_id" gorm:"type:uuid;primaryKey"`
	SkillName    string    `json:"skill_name" gorm:"type:varchar(100);primaryKey"`
	// HAPUS CreatedAt karena tidak ada di table database
}

func (ExperienceSkill) TableName() string {
	return "experience_skills"
}

// ============================
// REQUEST MODELS
// ============================

type ExperienceRequest struct {
	Title            string                  `json:"title" binding:"required"`
	Company          string                  `json:"company" binding:"required"`
	Location         string                  `json:"location" binding:"required"`
	StartYears       string                  `json:"start_year" binding:"required"`
	EndYears         string                  `json:"end_year" binding:"required"`
	CurrentJob       bool                    `json:"current_job"`
	DisplayOrder     int                     `json:"display_order"`
	Responsibilities []ResponsibilityRequest `json:"responsibilities"`
	Skills           []SkillRequest          `json:"skills"`
}

type ExperienceUpdateRequest struct {
	Title            string                  `json:"title"`
	Company          string                  `json:"company"`
	Location         string                  `json:"location"`
	StartYears       string                  `json:"start_year"`
	EndYears         string                  `json:"end_year"`
	CurrentJob       bool                    `json:"current_job"`
	DisplayOrder     int                     `json:"display_order"`
	Responsibilities []ResponsibilityRequest `json:"responsibilities"`
	Skills           []SkillRequest          `json:"skills"`
}

type ResponsibilityRequest struct {
	Description  string `json:"description" binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type SkillRequest struct {
	SkillName string `json:"skill_name" binding:"required"`
}

// ============================
// RESPONSE MODELS
// ============================

type ExperienceResponse struct {
	ID               uuid.UUID                `json:"id"`
	Title            string                   `json:"title"`
	Company          string                   `json:"company"`
	Location         string                   `json:"location"`
	StartYears       string                   `json:"start_year"`
	EndYears         string                   `json:"end_year"`
	CurrentJob       bool                     `json:"current_job"`
	DisplayOrder     int                      `json:"display_order"`
	Responsibilities []ResponsibilityResponse `json:"responsibilities"`
	Skills           []SkillResponse          `json:"skills"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type ResponsibilityResponse struct {
	ID           uuid.UUID `json:"id"`
	ExperienceID uuid.UUID `json:"experience_id"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type SkillResponse struct {
	ExperienceID uuid.UUID `json:"experience_id"`
	SkillName    string    `json:"skill_name"`
	// HAPUS CreatedAt karena tidak ada di table database
}
