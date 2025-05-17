package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type GroupService interface {
	CreateGroup(name, description string) (*models.Group, error)
	GetGroupByID(id uint) (*models.Group, error)
	UpdateGroup(id uint, name, description string) (*models.Group, error)
	DeleteGroup(id uint) error
	ListGroups() ([]models.Group, error)
}

type groupService struct {
	groupRepo   repositories.GroupRepository
	redisClient *redis.Client
}

func NewGroupService(gRepo repositories.GroupRepository, redisClient *redis.Client) GroupService {
	return &groupService{groupRepo: gRepo, redisClient: redisClient}
}

func (s *groupService) CreateGroup(name, description string) (*models.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	existingGroup, err := s.groupRepo.GetGroupByName(name)
	if err == nil && existingGroup != nil {
		return nil, fmt.Errorf("group already exists")
	}

	group := &models.Group{
		Name:        name,
		Description: description,
	}

	if err := s.groupRepo.CreateGroup(group); err != nil {
		log.Printf("Failed to create group %s: %v", name, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache("groups:*")

	return group, nil
}

func (s *groupService) GetGroupByID(id uint) (*models.Group, error) {
	cacheKey := fmt.Sprintf("group:%d", id)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var group models.Group
		if err := json.Unmarshal([]byte(cached), &group); err == nil {
			log.Printf("Cache hit for group:%d", id)
			return &group, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for group:%d: %v", id, err)
	}

	group, err := s.groupRepo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(group)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for group:%d: %v", id, err)
		}
	}

	return group, nil
}

func (s *groupService) UpdateGroup(id uint, name, description string) (*models.Group, error) {
	group, err := s.groupRepo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		existingGroup, err := s.groupRepo.GetGroupByName(name)
		if err == nil && existingGroup != nil && existingGroup.ID != id {
			return nil, fmt.Errorf("group name already exists")
		}
		group.Name = name
	}

	if description != "" {
		group.Description = description
	}

	if err := s.groupRepo.UpdateGroup(group); err != nil {
		log.Printf("Failed to update group %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("group:%d", id))
	s.invalidateCache("groups:*")

	return group, nil
}

func (s *groupService) DeleteGroup(id uint) error {
	err := s.groupRepo.DeleteGroup(id)
	if err != nil {
		log.Printf("Failed to delete group %d: %v", id, err)
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("group:%d", id))
	s.invalidateCache("groups:*")

	return nil
}

func (s *groupService) ListGroups() ([]models.Group, error) {
	cacheKey := "groups:all"

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var groups []models.Group
		if err := json.Unmarshal([]byte(cached), &groups); err == nil {
			log.Printf("Cache hit for groups:all")
			return groups, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for groups:all: %v", err)
	}

	groups, err := s.groupRepo.ListGroups()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(groups)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for groups:all: %v", err)
		}
	}

	return groups, nil
}

func (s *groupService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to invalidate cache for pattern %s: %v", pattern, err)
		return
	}

	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		}
	}
}
