package projectmodel

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title        string    `json:"title" gorm:"column:title;type:varchar(200);not null"`
	Description  string    `json:"description" gorm:"column:description;type:text;not null"`
	ImageURL     string    `json:"image_url" gorm:"column:image_url;type:varchar(500)"`
	DemoURL      string    `json:"demo_url" gorm:"column:demo_url;type:varchar(500)"`
	CodeURL      string    `json:"code_url" gorm:"column:code_url;type:varchar(500);not null"`
	DisplayOrder int       `json:"display_order" gorm:"column:display_order;type:integer;not null;default:0"`
	IsFeatured   bool      `json:"is_featured" gorm:"column:is_featured;type:boolean;default:false"`
	Status       string    `json:"status" gorm:"column:status;type:varchar(20);default:'published'"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Relations
	Tags []ProjectTag `json:"tags,omitempty" gorm:"many2many:project_tag_relations;"`
}

type ProjectForm struct {
	Title        string `form:"title" binding:"required"`
	Description  string `form:"description" binding:"required"`
	CodeURL      string `form:"code_url" binding:"required"`
	DemoURL      string `form:"demo_url"`
	Status       string `form:"status"`
	DisplayOrder int    `form:"display_order"`
	IsFeatured   bool   `form:"is_featured"`
}

type ProjectUpdateForm struct {
	Title        string `form:"title"`
	Description  string `form:"description"`
	DemoURL      string `form:"demo_url"`
	CodeURL      string `form:"code_url"`
	DisplayOrder int    `form:"display_order"`
	IsFeatured   bool   `form:"is_featured"`
	Status       string `form:"status"`
}

type ProjectTag struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(50);unique;not null"`
	Color     string    `json:"color" gorm:"column:color;type:varchar(7)"` // HEX color code
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

type TagResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

func (Project) TableName() string {
	return "portfolio_projects"
}

func (ProjectTag) TableName() string {
	return "project_tags"
}
