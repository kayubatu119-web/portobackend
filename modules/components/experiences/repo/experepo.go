package repo

import (
	"database/sql"
	"gintugas/modules/components/experiences/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExperiencesRepository interface {
	CreateExperienceWithRelations(experience *model.Experience) error
	GetExperienceByIDWithRelations(experienceID uuid.UUID) (*model.Experience, error)
	UpdateExperienceWithRelations(experience *model.Experience) error
	DeleteExperienceWithRelations(experienceID uuid.UUID) error
	GetAllExperiencesWithRelations() ([]model.Experience, error)
}

type experienceRepository struct {
	db *gorm.DB
}

func NewExpeGormRepository(db *gorm.DB) ExperiencesRepository {
	return &experienceRepository{
		db: db,
	}
}

type DbExperienceRepository interface {
	GetAllExperience() (result []model.Experience, err error)
	GetAllExperiencesWithRelations() ([]model.Experience, error)
}

type dbExperienceRepository struct {
	db *sql.DB
}

func NewDbExpeRepository(db *sql.DB) DbExperienceRepository {
	return &dbExperienceRepository{db: db}
}

// ============================
// GORM REPOSITORY IMPLEMENTATION
// ============================

func (r *experienceRepository) CreateExperienceWithRelations(experience *model.Experience) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create main experience
		if err := tx.Create(experience).Error; err != nil {
			return err
		}

		// Create responsibilities
		for i := range experience.Responsibilities {
			experience.Responsibilities[i].ExperienceID = experience.ID
			experience.Responsibilities[i].ID = uuid.Nil
		}
		if len(experience.Responsibilities) > 0 {
			if err := tx.Create(&experience.Responsibilities).Error; err != nil {
				return err
			}
		}

		// Create skills - PERBAIKAN: Gunakan Clauses untuk handle duplicate
		for i := range experience.Skills {
			experience.Skills[i].ExperienceID = experience.ID
		}
		if len(experience.Skills) > 0 {
			// Gunakan OnConflict untuk ignore duplicate keys
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "experience_id"}, {Name: "skill_name"}},
				DoNothing: true, // Ignore duplicate, tidak error
			}).Create(&experience.Skills).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *experienceRepository) GetExperienceByIDWithRelations(experienceID uuid.UUID) (*model.Experience, error) {
	var experience model.Experience
	err := r.db.
		Preload("Responsibilities", func(db *gorm.DB) *gorm.DB {
			return db.Order("experience_responsibilities.display_order ASC")
		}).
		Preload("Skills").
		Where("id = ?", experienceID).
		First(&experience).Error
	if err != nil {
		return nil, err
	}
	return &experience, nil
}

func (r *experienceRepository) UpdateExperienceWithRelations(experience *model.Experience) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update main experience
		if err := tx.Save(experience).Error; err != nil {
			return err
		}

		// Delete existing responsibilities and create new ones
		if err := tx.Where("experience_id = ?", experience.ID).Delete(&model.ExperienceResponsibility{}).Error; err != nil {
			return err
		}
		if len(experience.Responsibilities) > 0 {
			for i := range experience.Responsibilities {
				experience.Responsibilities[i].ExperienceID = experience.ID
				experience.Responsibilities[i].ID = uuid.Nil
			}
			if err := tx.Create(&experience.Responsibilities).Error; err != nil {
				return err
			}
		}

		// Delete existing skills and create new ones
		if err := tx.Where("experience_id = ?", experience.ID).Delete(&model.ExperienceSkill{}).Error; err != nil {
			return err
		}
		if len(experience.Skills) > 0 {
			for i := range experience.Skills {
				experience.Skills[i].ExperienceID = experience.ID
			}
			// Gunakan OnConflict untuk handle duplicate
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "experience_id"}, {Name: "skill_name"}},
				DoNothing: true,
			}).Create(&experience.Skills).Error; err != nil {
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

func (r *experienceRepository) GetAllExperiencesWithRelations() ([]model.Experience, error) {
	var experiences []model.Experience
	err := r.db.
		Preload("Responsibilities", func(db *gorm.DB) *gorm.DB {
			return db.Order("experience_responsibilities.display_order ASC")
		}).
		Preload("Skills").
		Order("display_order ASC, created_at DESC").
		Find(&experiences).Error
	if err != nil {
		return nil, err
	}
	return experiences, nil
}

// ============================
// SQL REPOSITORY IMPLEMENTATION
// ============================

func (r *dbExperienceRepository) GetAllExperience() ([]model.Experience, error) {
	query := "SELECT id, title, company, location, start_year, end_year, current_job, display_order, created_at, updated_at FROM portfolio_experiences ORDER BY display_order ASC, created_at DESC"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []model.Experience
	for rows.Next() {
		var experience model.Experience
		err := rows.Scan(
			&experience.ID,
			&experience.Title,
			&experience.Company,
			&experience.Location,
			&experience.StartYears,
			&experience.EndYears,
			&experience.CurrentJob,
			&experience.DisplayOrder,
			&experience.CreatedAt,
			&experience.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, experience)
	}

	return experiences, nil
}

func (r *dbExperienceRepository) GetAllExperiencesWithRelations() ([]model.Experience, error) {
	// Get all experiences
	experiences, err := r.GetAllExperience()
	if err != nil {
		return nil, err
	}

	// Get experience IDs for batch loading
	experienceIDs := make([]uuid.UUID, len(experiences))
	for i, exp := range experiences {
		experienceIDs[i] = exp.ID
	}

	if len(experienceIDs) == 0 {
		return experiences, nil
	}

	// Load responsibilities
	responsibilities, err := r.getResponsibilitiesByExperienceIDs(experienceIDs)
	if err != nil {
		return nil, err
	}

	// Load skills
	skills, err := r.getSkillsByExperienceIDs(experienceIDs)
	if err != nil {
		return nil, err
	}

	// Map relationships to experiences
	for i := range experiences {
		expID := experiences[i].ID
		experiences[i].Responsibilities = responsibilities[expID]
		experiences[i].Skills = skills[expID]
	}

	return experiences, nil
}

func (r *dbExperienceRepository) getResponsibilitiesByExperienceIDs(experienceIDs []uuid.UUID) (map[uuid.UUID][]model.ExperienceResponsibility, error) {
	query := `SELECT id, experience_id, description, display_order, created_at 
			  FROM experience_responsibilities 
			  WHERE experience_id = ANY($1) 
			  ORDER BY display_order ASC`
	rows, err := r.db.Query(query, experienceIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	responsibilities := make(map[uuid.UUID][]model.ExperienceResponsibility)
	for rows.Next() {
		var resp model.ExperienceResponsibility
		err := rows.Scan(
			&resp.ID,
			&resp.ExperienceID,
			&resp.Description,
			&resp.DisplayOrder,
			&resp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		responsibilities[resp.ExperienceID] = append(responsibilities[resp.ExperienceID], resp)
	}

	return responsibilities, nil
}

func (r *dbExperienceRepository) getSkillsByExperienceIDs(experienceIDs []uuid.UUID) (map[uuid.UUID][]model.ExperienceSkill, error) {
	query := `SELECT experience_id, skill_name 
			  FROM experience_skills 
			  WHERE experience_id = ANY($1)`
	rows, err := r.db.Query(query, experienceIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make(map[uuid.UUID][]model.ExperienceSkill)
	for rows.Next() {
		var skill model.ExperienceSkill
		err := rows.Scan(
			&skill.ExperienceID,
			&skill.SkillName,
		)
		if err != nil {
			return nil, err
		}
		skills[skill.ExperienceID] = append(skills[skill.ExperienceID], skill)
	}

	return skills, nil
}
