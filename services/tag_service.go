package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type TagService interface {
	CreateTag(name string) (*models.Tag, error)
	GetTagByID(id uint) (*models.Tag, error)
	GetTagByName(name string) (*models.Tag, error)
	UpdateTag(id uint, name string) (*models.Tag, error)
	DeleteTag(id uint) error
	ListTags() ([]models.Tag, error)
}

type tagService struct {
	tagRepo repositories.TagRepository
}

func NewTagService(tRepo repositories.TagRepository) TagService {
	return &tagService{tagRepo: tRepo}
}

func (s *tagService) CreateTag(name string) (*models.Tag, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	// Kiểm tra xem tag đã tồn tại chưa
	existingTag, err := s.tagRepo.GetTagByName(name)
	if err == nil && existingTag != nil {
		return nil, errors.New("tag already exists")
	}

	tag := &models.Tag{
		Name: name,
	}

	if err := s.tagRepo.CreateTag(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) GetTagByID(id uint) (*models.Tag, error) {
	return s.tagRepo.GetTagByID(id)
}

func (s *tagService) GetTagByName(name string) (*models.Tag, error) {
	return s.tagRepo.GetTagByName(name)
}

func (s *tagService) UpdateTag(id uint, name string) (*models.Tag, error) {
	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		// Kiểm tra xem tên tag đã tồn tại chưa
		existingTag, err := s.tagRepo.GetTagByName(name)
		if err == nil && existingTag != nil && existingTag.ID != id {
			return nil, errors.New("tag name already exists")
		}
		tag.Name = name
	}

	if err := s.tagRepo.UpdateTag(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) DeleteTag(id uint) error {
	return s.tagRepo.DeleteTag(id)
}

func (s *tagService) ListTags() ([]models.Tag, error) {
	return s.tagRepo.ListTags()
}
