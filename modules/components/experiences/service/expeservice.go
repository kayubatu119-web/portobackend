package service

import (
	"errors"
	"gintugas/modules/components/experiences/model"
	"gintugas/modules/components/experiences/repo"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExperiencesService interface {
	CreateExperienceWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error)
	GetExperienceByIDWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error)
	UpdateExperienceWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error)
	DeleteExperienceWithRelations(ctx *gin.Context) error
	GetAllExperiencesWithRelations(ctx *gin.Context) ([]model.ExperienceResponse, error)
}

type experiencesService struct {
	experienceRepo repo.ExperiencesRepository
}

func NewExpeService(experienceRepo repo.ExperiencesRepository) ExperiencesService {
	return &experiencesService{
		experienceRepo: experienceRepo,
	}
}

func (s *experiencesService) CreateExperienceWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error) {
	var experienceReq model.ExperienceRequest
	if err := ctx.ShouldBindJSON(&experienceReq); err != nil {
		return nil, err
	}

	experience := &model.ExperienceWithRelations{
		Experience: model.Experience{
			Title:        experienceReq.Title,
			Company:      experienceReq.Company,
			Location:     experienceReq.Location,
			StartYears:   experienceReq.StartYears,
			EndYears:     experienceReq.EndYears,
			CurrentJob:   experienceReq.CurrentJob,
			DisplayOrder: experienceReq.DisplayOrder,
		},
	}

	// Convert responsibilities
	for _, respReq := range experienceReq.Responsibilities {
		experience.Responsibilities = append(experience.Responsibilities, model.ExperienceResponsibility{
			Description:  respReq.Description,
			DisplayOrder: respReq.DisplayOrder,
		})
	}

	// Convert skills
	for _, skillReq := range experienceReq.Skills {
		experience.Skills = append(experience.Skills, model.ExperienceSkill{
			SkillName: skillReq.SkillName,
		})
	}

	if err := s.experienceRepo.CreateExperienceWithRelations(experience); err != nil {
		return nil, err
	}

	return s.convertToResponse(experience), nil
}

func (s *experiencesService) GetExperienceByIDWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error) {
	experienceID := ctx.Param("id")
	experienceUUID, err := uuid.Parse(experienceID)
	if err != nil {
		return nil, errors.New("format experience ID tidak valid")
	}

	experience, err := s.experienceRepo.GetExperienceByIDWithRelations(experienceUUID)
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(experience), nil
}

func (s *experiencesService) UpdateExperienceWithRelations(ctx *gin.Context) (*model.ExperienceResponse, error) {
	experienceID := ctx.Param("id")
	experienceUUID, err := uuid.Parse(experienceID)
	if err != nil {
		return nil, errors.New("format experience ID tidak valid")
	}

	// Get existing experience
	existingExperience, err := s.experienceRepo.GetExperienceByIDWithRelations(experienceUUID)
	if err != nil {
		return nil, err
	}

	var experienceReq model.ExperienceUpdateRequest
	if err := ctx.ShouldBindJSON(&experienceReq); err != nil {
		return nil, err
	}

	// Update main fields
	if experienceReq.Title != "" {
		existingExperience.Title = experienceReq.Title
	}
	if experienceReq.Company != "" {
		existingExperience.Company = experienceReq.Company
	}
	if experienceReq.Location != "" {
		existingExperience.Location = experienceReq.Location
	}
	if experienceReq.StartYears != "" {
		existingExperience.StartYears = experienceReq.StartYears
	}
	if experienceReq.EndYears != "" {
		existingExperience.EndYears = experienceReq.EndYears
	}
	existingExperience.CurrentJob = experienceReq.CurrentJob
	existingExperience.DisplayOrder = experienceReq.DisplayOrder
	existingExperience.UpdatedAt = time.Now()

	// Update responsibilities
	existingExperience.Responsibilities = nil
	for _, respReq := range experienceReq.Responsibilities {
		existingExperience.Responsibilities = append(existingExperience.Responsibilities, model.ExperienceResponsibility{
			Description:  respReq.Description,
			DisplayOrder: respReq.DisplayOrder,
		})
	}

	// Update skills
	existingExperience.Skills = nil
	for _, skillReq := range experienceReq.Skills {
		existingExperience.Skills = append(existingExperience.Skills, model.ExperienceSkill{
			SkillName: skillReq.SkillName,
		})
	}

	if err := s.experienceRepo.UpdateExperienceWithRelations(existingExperience); err != nil {
		return nil, err
	}

	return s.convertToResponse(existingExperience), nil
}

func (s *experiencesService) DeleteExperienceWithRelations(ctx *gin.Context) error {
	experienceID := ctx.Param("id")
	experienceUUID, err := uuid.Parse(experienceID)
	if err != nil {
		return errors.New("format experience ID tidak valid")
	}

	return s.experienceRepo.DeleteExperienceWithRelations(experienceUUID)
}

func (s *experiencesService) GetAllExperiencesWithRelations(ctx *gin.Context) ([]model.ExperienceResponse, error) {
	experiences, err := s.experienceRepo.GetAllExperiencesWithRelations()
	if err != nil {
		return nil, err
	}

	var responses []model.ExperienceResponse
	for _, exp := range experiences {
		responses = append(responses, *s.convertToResponse(&exp))
	}

	return responses, nil
}

func (s *experiencesService) convertToResponse(experience *model.ExperienceWithRelations) *model.ExperienceResponse {
	// Convert responsibilities
	var respResponses []model.ResponsibilityResponse
	for _, resp := range experience.Responsibilities {
		respResponses = append(respResponses, model.ResponsibilityResponse{
			ID:           resp.ID,
			ExperienceID: resp.ExperienceID,
			Description:  resp.Description,
			DisplayOrder: resp.DisplayOrder,
			CreatedAt:    resp.CreatedAt,
		})
	}

	// Convert skills
	var skillResponses []model.SkillResponse
	for _, skill := range experience.Skills {
		skillResponses = append(skillResponses, model.SkillResponse{
			ExperienceID: skill.ExperienceID,
			SkillName:    skill.SkillName,
		})
	}

	return &model.ExperienceResponse{
		ID:               experience.ID,
		Title:            experience.Title,
		Company:          experience.Company,
		Location:         experience.Location,
		StartYears:       experience.StartYears,
		EndYears:         experience.EndYears,
		CurrentJob:       experience.CurrentJob,
		DisplayOrder:     experience.DisplayOrder,
		Responsibilities: respResponses,
		Skills:           skillResponses,
		CreatedAt:        experience.CreatedAt,
		UpdatedAt:        experience.UpdatedAt,
	}
}
