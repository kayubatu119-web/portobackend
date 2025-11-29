package repo

import (
	"gintugas/modules/components/experiences/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExperiencesRepository interface {
	CreateExperienceWithRelations(experience *model.ExperienceWithRelations) error
	GetExperienceByIDWithRelations(experienceID uuid.UUID) (*model.ExperienceWithRelations, error)
	UpdateExperienceWithRelations(experience *model.ExperienceWithRelations) error
	DeleteExperienceWithRelations(experienceID uuid.UUID) error
	GetAllExperiencesWithRelations() ([]model.ExperienceWithRelations, error)
}

type experienceRepository struct {
	db *gorm.DB
}

func NewExpeGormRepository(db *gorm.DB) ExperiencesRepository {
	return &experienceRepository{
		db: db,
	}
}

func (r *experienceRepository) CreateExperienceWithRelations(experience *model.ExperienceWithRelations) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create main experience first
		if err := tx.Create(&experience.Experience).Error; err != nil {
			return err
		}

		// Create responsibilities
		if len(experience.Responsibilities) > 0 {
			responsibilities := make([]model.ExperienceResponsibility, len(experience.Responsibilities))
			for i, resp := range experience.Responsibilities {
				responsibilities[i] = model.ExperienceResponsibility{
					ExperienceID: experience.ID,
					Description:  resp.Description,
					DisplayOrder: resp.DisplayOrder,
				}
			}
			if err := tx.Create(&responsibilities).Error; err != nil {
				return err
			}
			experience.Responsibilities = responsibilities
		}

		// Create skills
		if len(experience.Skills) > 0 {
			skills := make([]model.ExperienceSkill, len(experience.Skills))
			for i, skill := range experience.Skills {
				skills[i] = model.ExperienceSkill{
					ExperienceID: experience.ID,
					SkillName:    skill.SkillName,
				}
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "experience_id"}, {Name: "skill_name"}},
				DoNothing: true,
			}).Create(&skills).Error; err != nil {
				return err
			}
			experience.Skills = skills
		}

		return nil
	})
}

func (r *experienceRepository) GetExperienceByIDWithRelations(experienceID uuid.UUID) (*model.ExperienceWithRelations, error) {
	var experience model.Experience
	err := r.db.Where("id = ?", experienceID).First(&experience).Error
	if err != nil {
		return nil, err
	}

	// Load responsibilities manually
	var responsibilities []model.ExperienceResponsibility
	err = r.db.Where("experience_id = ?", experienceID).
		Order("display_order ASC").
		Find(&responsibilities).Error
	if err != nil {
		return nil, err
	}

	// Load skills manually
	var skills []model.ExperienceSkill
	err = r.db.Where("experience_id = ?", experienceID).Find(&skills).Error
	if err != nil {
		return nil, err
	}

	return &model.ExperienceWithRelations{
		Experience:       experience,
		Responsibilities: responsibilities,
		Skills:           skills,
	}, nil
}

func (r *experienceRepository) GetAllExperiencesWithRelations() ([]model.ExperienceWithRelations, error) {
	// Get all experiences
	var experiences []model.Experience
	err := r.db.Order("display_order ASC, created_at DESC").Find(&experiences).Error
	if err != nil {
		return nil, err
	}

	if len(experiences) == 0 {
		return []model.ExperienceWithRelations{}, nil
	}

	// Get experience IDs
	experienceIDs := make([]uuid.UUID, len(experiences))
	for i, exp := range experiences {
		experienceIDs[i] = exp.ID
	}

	// Load all responsibilities for these experiences
	var allResponsibilities []model.ExperienceResponsibility
	err = r.db.Where("experience_id IN (?)", experienceIDs).
		Order("experience_id, display_order ASC").
		Find(&allResponsibilities).Error
	if err != nil {
		return nil, err
	}

	// Load all skills for these experiences
	var allSkills []model.ExperienceSkill
	err = r.db.Where("experience_id IN (?)", experienceIDs).Find(&allSkills).Error
	if err != nil {
		return nil, err
	}

	// Group responsibilities by experience ID
	responsibilitiesByExp := make(map[uuid.UUID][]model.ExperienceResponsibility)
	for _, resp := range allResponsibilities {
		responsibilitiesByExp[resp.ExperienceID] = append(responsibilitiesByExp[resp.ExperienceID], resp)
	}

	// Group skills by experience ID
	skillsByExp := make(map[uuid.UUID][]model.ExperienceSkill)
	for _, skill := range allSkills {
		skillsByExp[skill.ExperienceID] = append(skillsByExp[skill.ExperienceID], skill)
	}

	// Combine everything
	var result []model.ExperienceWithRelations
	for _, exp := range experiences {
		result = append(result, model.ExperienceWithRelations{
			Experience:       exp,
			Responsibilities: responsibilitiesByExp[exp.ID],
			Skills:           skillsByExp[exp.ID],
		})
	}

	return result, nil
}

func (r *experienceRepository) UpdateExperienceWithRelations(experience *model.ExperienceWithRelations) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update main experience
		if err := tx.Save(&experience.Experience).Error; err != nil {
			return err
		}

		// Delete existing responsibilities
		if err := tx.Where("experience_id = ?", experience.ID).Delete(&model.ExperienceResponsibility{}).Error; err != nil {
			return err
		}

		// Create new responsibilities
		if len(experience.Responsibilities) > 0 {
			for i := range experience.Responsibilities {
				experience.Responsibilities[i].ExperienceID = experience.ID
				experience.Responsibilities[i].ID = uuid.Nil // Let DB generate
			}
			if err := tx.Create(&experience.Responsibilities).Error; err != nil {
				return err
			}
		}

		// Delete existing skills
		if err := tx.Where("experience_id = ?", experience.ID).Delete(&model.ExperienceSkill{}).Error; err != nil {
			return err
		}

		// Create new skills
		if len(experience.Skills) > 0 {
			for i := range experience.Skills {
				experience.Skills[i].ExperienceID = experience.ID
			}
			if err := tx.Create(&experience.Skills).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *experienceRepository) DeleteExperienceWithRelations(experienceID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete skills
		if err := tx.Where("experience_id = ?", experienceID).Delete(&model.ExperienceSkill{}).Error; err != nil {
			return err
		}

		// Delete responsibilities
		if err := tx.Where("experience_id = ?", experienceID).Delete(&model.ExperienceResponsibility{}).Error; err != nil {
			return err
		}

		// Delete main experience
		return tx.Where("id = ?", experienceID).Delete(&model.Experience{}).Error
	})
}
