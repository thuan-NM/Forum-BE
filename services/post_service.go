package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
)

type PostService interface {
	CreatePost(content string, userID uint, groupID uint, status models.PostStatus) (*models.Post, error)
	GetPostByID(id uint) (*models.Post, error)
	DeletePost(id uint) error
	UpdatePost(id uint, content string, status models.PostStatus) (*models.Post, error)
	ListPosts(filters map[string]interface{}) ([]models.Post, error)
}

type postService struct {
	postRepo repositories.PostRepository
}

func NewPostService(postRepo repositories.PostRepository) PostService {
	return &postService{postRepo}
}

func (s *postService) CreatePost(content string, userID uint, groupID uint, status models.PostStatus) (*models.Post, error) {
	post := &models.Post{
		Content: content,
		UserID:  userID,
		GroupID: groupID,
		Status:  status,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) GetPostByID(id uint) (*models.Post, error) {
	return s.postRepo.GetPostByID(id)
}

func (s *postService) DeletePost(id uint) error {
	return s.postRepo.DeletePost(id)
}

func (s *postService) UpdatePost(id uint, content string, status models.PostStatus) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		return nil, err
	}
	post.Content = content
	post.Status = status

	if err := s.postRepo.UpdatePost(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) ListPosts(filters map[string]interface{}) ([]models.Post, error) {
	return s.postRepo.List(filters)
}
