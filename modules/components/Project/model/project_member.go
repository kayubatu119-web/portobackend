package projectmodel

import "github.com/google/uuid"

type ProjectTagRelation struct {
	ProjectID uuid.UUID  `json:"project_id" gorm:"type:uuid;primaryKey"`
	TagID     uuid.UUID  `json:"tag_id" gorm:"type:uuid;primaryKey"`
	Project   Project    `json:"project,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
	Tag       ProjectTag `json:"tag,omitempty" gorm:"foreignKey:TagID;references:ID"` // Ubah dari Tags ke Tag, dan UserID ke TagID
}

func (ProjectTagRelation) TableName() string {
	return "project_tag_relations"
}
