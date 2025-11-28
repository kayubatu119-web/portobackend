package projectrepo

import (
	model "gintugas/modules/components/Project/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectMemberRepo struct {
	DB *gorm.DB
}

func NewProjectMemberRepo(db *gorm.DB) *ProjectMemberRepo {
	return &ProjectMemberRepo{DB: db}
}

func (r *ProjectMemberRepo) AddTag(projectID, tagID string) error {
	// Parse string ke UUID
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return err
	}

	tID, err := uuid.Parse(tagID)
	if err != nil {
		return err
	}

	relation := model.ProjectTagRelation{
		ProjectID: pID,
		TagID:     tID,
	}

	return r.DB.Create(&relation).Error
}

func (r *ProjectMemberRepo) RemoveTag(projectID, tagID string) error {
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return err
	}

	tID, err := uuid.Parse(tagID)
	if err != nil {
		return err
	}

	return r.DB.Where("project_id = ? AND tag_id = ?", pID, tID).
		Delete(&model.ProjectTagRelation{}).Error
}

func (r *ProjectMemberRepo) GetProjectTags(projectID string) ([]model.ProjectTag, error) {
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	var tags []model.ProjectTag

	err = r.DB.
		Joins("JOIN project_tag_relations ON project_tag_relations.tag_id = project_tags.id").
		Where("project_tag_relations.project_id = ?", pID).
		Find(&tags).Error

	return tags, err
}

func (r *ProjectMemberRepo) IsProjectTag(projectID, tagID string) (bool, error) {
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return false, err
	}

	tID, err := uuid.Parse(tagID)
	if err != nil {
		return false, err
	}

	var count int64
	err = r.DB.Model(&model.ProjectTagRelation{}).
		Where("project_id = ? AND tag_id = ?", pID, tID).
		Count(&count).Error

	return count > 0, err
}

func (r *ProjectMemberRepo) GetProjectsByTag(tagID string) ([]model.Project, error) {
	tID, err := uuid.Parse(tagID)
	if err != nil {
		return nil, err
	}

	var projects []model.Project

	err = r.DB.
		Joins("JOIN project_tag_relations ON project_tag_relations.project_id = portfolio_projects.id").
		Where("project_tag_relations.tag_id = ?", tID).
		Preload("Tags").
		Find(&projects).Error

	return projects, err
}
