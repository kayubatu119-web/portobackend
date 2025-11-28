package model

import (
	"time"

	"github.com/google/uuid"
)

// ============================
// SKILLS MODEL
// ============================

type Skill struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string    `json:"name" gorm:"type:varchar(100);unique;not null"`
	Value        int       `json:"value" gorm:"type:integer;check:value >= 0 AND value <= 100"`
	IconURL      string    `json:"icon_url" gorm:"type:varchar(500)"`
	Category     string    `json:"category" gorm:"type:varchar(50)"` // programming, framework, tool
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	IsFeatured   bool      `json:"is_featured" gorm:"type:boolean;default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Skill) TableName() string {
	return "portfolio_skills"
}

type SkillForm struct {
	Name         string `form:"name" binding:"required"`
	Value        int    `form:"value" binding:"required,min=0,max=100"`
	Category     string `form:"category"`
	DisplayOrder int    `form:"display_order"`
	IsFeatured   bool   `form:"is_featured"`
}

type SkillRequest struct {
	Name         string `json:"name" binding:"required"`
	Value        int    `json:"value" binding:"required,min=0,max=100"`
	IconURL      string `json:"icon_url"`
	Category     string `json:"category"`
	DisplayOrder int    `json:"display_order"`
	IsFeatured   bool   `json:"is_featured"`
}

type SkillUpdateRequest struct {
	Name         string `json:"name"`
	Value        int    `json:"value" binding:"omitempty,min=0,max=100"`
	IconURL      string `json:"icon_url"`
	Category     string `json:"category"`
	DisplayOrder int    `json:"display_order"`
	IsFeatured   bool   `json:"is_featured"`
}

type SkillResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Value        int       `json:"value"`
	IconURL      string    `json:"icon_url"`
	Category     string    `json:"category"`
	DisplayOrder int       `json:"display_order"`
	IsFeatured   bool      `json:"is_featured"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================
// CERTIFICATES MODEL
// ============================

type Certificate struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"type:varchar(200);not null"`
	ImageURL      string    `json:"image_url" gorm:"type:varchar(500);not null"`
	IssueDate     time.Time `json:"issue_date" gorm:"type:date"`
	Issuer        string    `json:"issuer" gorm:"type:varchar(150)"`
	CredentialURL string    `json:"credential_url" gorm:"type:varchar(500)"`
	DisplayOrder  int       `json:"display_order" gorm:"type:integer;default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Certificate) TableName() string {
	return "portfolio_certificates"
}

type CertificateForm struct {
	Name          string `form:"name" binding:"required"`
	IssueDate     string `form:"issue_date"` // Pakai string untuk form-data
	Issuer        string `form:"issuer"`
	CredentialURL string `form:"credential_url"`
	DisplayOrder  int    `form:"display_order"`
}

type CertificateRequest struct {
	Name          string    `json:"name" binding:"required"`
	ImageURL      string    `json:"image_url" binding:"required"`
	IssueDate     time.Time `json:"issue_date"`
	Issuer        string    `json:"issuer"`
	CredentialURL string    `json:"credential_url"`
	DisplayOrder  int       `json:"display_order"`
}

type CertificateUpdateRequest struct {
	Name          string    `json:"name"`
	ImageURL      string    `json:"image_url"`
	IssueDate     time.Time `json:"issue_date"`
	Issuer        string    `json:"issuer"`
	CredentialURL string    `json:"credential_url"`
	DisplayOrder  int       `json:"display_order"`
}

type CertificateResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ImageURL      string    `json:"image_url"`
	IssueDate     time.Time `json:"issue_date"`
	Issuer        string    `json:"issuer"`
	CredentialURL string    `json:"credential_url"`
	DisplayOrder  int       `json:"display_order"`
	CreatedAt     time.Time `json:"created_at"`
}

// ============================
// EDUCATION MODEL
// ============================

type Education struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	School       string                 `json:"school" gorm:"type:varchar(200);not null"`
	Major        string                 `json:"major" gorm:"type:varchar(200);not null"`
	StartYear    string                 `json:"start_year" gorm:"type:varchar(20)"`
	EndYear      string                 `json:"end_year" gorm:"type:varchar(20)"`
	Description  string                 `json:"description" gorm:"type:text"`
	Degree       string                 `json:"degree" gorm:"type:varchar(100)"` // S1, S2, SMA
	DisplayOrder int                    `json:"display_order" gorm:"type:integer;default:0"`
	Achievements []EducationAchievement `json:"achievements" gorm:"foreignKey:EducationID;references:ID"`
	CreatedAt    time.Time              `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Education) TableName() string {
	return "portfolio_education"
}

type EducationAchievement struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EducationID  uuid.UUID `json:"education_id" gorm:"type:uuid;not null"`
	Achievement  string    `json:"achievement" gorm:"type:text;not null"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (EducationAchievement) TableName() string {
	return "education_achievements"
}

type AchievementRequest struct {
	Achievement  string `json:"achievement" binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type EducationRequest struct {
	School       string               `json:"school" binding:"required"`
	Major        string               `json:"major" binding:"required"`
	StartYear    string               `json:"start_year"`
	EndYear      string               `json:"end_year"`
	Description  string               `json:"description"`
	Degree       string               `json:"degree"`
	DisplayOrder int                  `json:"display_order"`
	Achievements []AchievementRequest `json:"achievements"`
}

type EducationUpdateRequest struct {
	School       string               `json:"school"`
	Major        string               `json:"major"`
	StartYear    string               `json:"start_year"`
	EndYear      string               `json:"end_year"`
	Description  string               `json:"description"`
	Degree       string               `json:"degree"`
	DisplayOrder int                  `json:"display_order"`
	Achievements []AchievementRequest `json:"achievements"`
}

type AchievementResponse struct {
	ID           uuid.UUID `json:"id"`
	EducationID  uuid.UUID `json:"education_id"`
	Achievement  string    `json:"achievement"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type EducationResponse struct {
	ID           uuid.UUID             `json:"id"`
	School       string                `json:"school"`
	Major        string                `json:"major"`
	StartYear    string                `json:"start_year"`
	EndYear      string                `json:"end_year"`
	Description  string                `json:"description"`
	Degree       string                `json:"degree"`
	DisplayOrder int                   `json:"display_order"`
	Achievements []AchievementResponse `json:"achievements"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// ============================
// TESTIMONIALS MODEL
// ============================

type Testimonial struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null"`
	Title        string    `json:"title" gorm:"type:varchar(150);not null"`
	Message      string    `json:"message" gorm:"type:text;not null"`
	AvatarURL    string    `json:"avatar_url" gorm:"type:varchar(500)"`
	Rating       int       `json:"rating" gorm:"type:integer;check:rating >= 1 AND rating <= 5"`
	IsFeatured   bool      `json:"is_featured" gorm:"type:boolean;default:false"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	Status       string    `json:"status" gorm:"type:varchar(20);default:'approved'"` // pending, approved, rejected
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Testimonial) TableName() string {
	return "portfolio_testimonials"
}

type TestimonialRequest struct {
	Name         string `json:"name" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Message      string `json:"message" binding:"required"`
	AvatarURL    string `json:"avatar_url"`
	Rating       int    `json:"rating" binding:"required,min=1,max=5"`
	IsFeatured   bool   `json:"is_featured"`
	DisplayOrder int    `json:"display_order"`
	Status       string `json:"status"`
}

type TestimonialUpdateRequest struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Message      string `json:"message"`
	AvatarURL    string `json:"avatar_url"`
	Rating       int    `json:"rating" binding:"omitempty,min=1,max=5"`
	IsFeatured   bool   `json:"is_featured"`
	DisplayOrder int    `json:"display_order"`
	Status       string `json:"status"`
}

type TestimonialResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	AvatarURL    string    `json:"avatar_url"`
	Rating       int       `json:"rating"`
	IsFeatured   bool      `json:"is_featured"`
	DisplayOrder int       `json:"display_order"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ============================
// BLOG MODELS
// ============================

type BlogPost struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title         string    `json:"title" gorm:"type:varchar(200);not null"`
	Content       string    `json:"content" gorm:"type:text"`
	Excerpt       string    `json:"excerpt" gorm:"type:text"`
	Slug          string    `json:"slug" gorm:"type:varchar(200);unique;not null"`
	FeaturedImage string    `json:"featured_image" gorm:"type:varchar(500)"`
	PublishDate   time.Time `json:"publish_date" gorm:"type:date"`
	Status        string    `json:"status" gorm:"type:varchar(20);default:'draft'"` // draft, published, archived
	ViewCount     int       `json:"view_count" gorm:"type:integer;default:0"`
	Tags          []BlogTag `json:"tags" gorm:"many2many:blog_post_tags;joinForeignKey:PostID;joinReferences:TagID"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (BlogPost) TableName() string {
	return "portfolio_blog_posts"
}

type BlogTag struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"type:varchar(50);unique;not null"`
	Posts     []BlogPost `json:"-" gorm:"many2many:blog_post_tags;joinForeignKey:TagID;joinReferences:PostID"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (BlogTag) TableName() string {
	return "blog_tags"
}

type BlogPostTag struct {
	PostID uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey;column:post_id"`
	TagID  uuid.UUID `json:"tag_id" gorm:"type:uuid;primaryKey;column:tag_id"`
}

func (BlogPostTag) TableName() string {
	return "blog_post_tags"
}

type TagRequest struct {
	Name string `json:"name" binding:"required"`
}

type BlogPostRequest struct {
	Title         string       `json:"title" binding:"required"`
	Content       string       `json:"content"`
	Excerpt       string       `json:"excerpt"`
	Slug          string       `json:"slug" binding:"required"`
	FeaturedImage string       `json:"featured_image"`
	PublishDate   time.Time    `json:"publish_date"`
	Status        string       `json:"status"`
	Tags          []TagRequest `json:"tags"`
}

type BlogPostUpdateRequest struct {
	Title         string       `json:"title"`
	Content       string       `json:"content"`
	Excerpt       string       `json:"excerpt"`
	Slug          string       `json:"slug"`
	FeaturedImage string       `json:"featured_image"`
	PublishDate   time.Time    `json:"publish_date"`
	Status        string       `json:"status"`
	Tags          []TagRequest `json:"tags"`
}

type TagResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type BlogPostResponse struct {
	ID            uuid.UUID     `json:"id"`
	Title         string        `json:"title"`
	Content       string        `json:"content"`
	Excerpt       string        `json:"excerpt"`
	Slug          string        `json:"slug"`
	FeaturedImage string        `json:"featured_image"`
	PublishDate   time.Time     `json:"publish_date"`
	Status        string        `json:"status"`
	ViewCount     int           `json:"view_count"`
	Tags          []TagResponse `json:"tags"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ============================
// SECTIONS MODEL
// ============================

type Section struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SectionID    string    `json:"section_id" gorm:"type:varchar(50);unique;not null"`
	Label        string    `json:"label" gorm:"type:varchar(100);not null"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	IsActive     bool      `json:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Section) TableName() string {
	return "portfolio_sections"
}

type SectionRequest struct {
	SectionID    string `json:"section_id" binding:"required"`
	Label        string `json:"label" binding:"required"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
}

type SectionResponse struct {
	ID           uuid.UUID `json:"id"`
	SectionID    string    `json:"section_id"`
	Label        string    `json:"label"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================
// SOCIAL LINKS MODEL
// ============================

type SocialLink struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Platform     string    `json:"platform" gorm:"type:varchar(50);unique;not null"`
	URL          string    `json:"url" gorm:"type:varchar(500);not null"`
	IconName     string    `json:"icon_name" gorm:"type:varchar(50)"`
	DisplayOrder int       `json:"display_order" gorm:"type:integer;default:0"`
	IsActive     bool      `json:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (SocialLink) TableName() string {
	return "portfolio_social_links"
}

type SocialLinkRequest struct {
	Platform     string `json:"platform" binding:"required"`
	URL          string `json:"url" binding:"required"`
	IconName     string `json:"icon_name"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
}

type SocialLinkResponse struct {
	ID           uuid.UUID `json:"id"`
	Platform     string    `json:"platform"`
	URL          string    `json:"url"`
	IconName     string    `json:"icon_name"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================
// SETTINGS MODEL
// ============================

type Setting struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key         string    `json:"key" gorm:"type:varchar(100);unique;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	DataType    string    `json:"data_type" gorm:"type:varchar(20);default:'string'"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (Setting) TableName() string {
	return "portfolio_settings"
}

type SettingRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
}

type SettingResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	DataType    string    `json:"data_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
