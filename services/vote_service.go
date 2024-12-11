package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

// VoteService định nghĩa các phương thức liên quan đến Vote
type VoteService interface {
	CastVote(userID uint, votableType string, votableID uint, voteType string) (*models.Vote, error)
	GetVoteByID(id uint) (*models.Vote, error)
	UpdateVote(id uint, voteType string) (*models.Vote, error)
	DeleteVote(id uint) error
	ListVotes() ([]models.Vote, error)
	GetVoteByUserAndVotable(userID uint, votableType string, votableID uint) (*models.Vote, error)
	GetVoteCount(votableType string, votableID uint) (int64, error)
}

type voteService struct {
	voteRepo repositories.VoteRepository
}

func NewVoteService(vRepo repositories.VoteRepository) VoteService {
	return &voteService{voteRepo: vRepo}
}

func (s *voteService) CastVote(userID uint, votableType string, votableID uint, voteType string) (*models.Vote, error) {
	// Kiểm tra valid votableType
	if votableType != "question" && votableType != "answer" && votableType != "comment" {
		return nil, errors.New("invalid votable type")
	}

	// Kiểm tra valid voteType
	if voteType != string(models.VoteUp) && voteType != string(models.VoteDown) {
		return nil, errors.New("invalid vote type")
	}

	// Kiểm tra xem người dùng đã vote cho đối tượng này chưa
	existingVote, err := s.voteRepo.GetVoteByUserAndVotable(userID, votableType, votableID)
	if err == nil && existingVote != nil {
		return nil, errors.New("user has already voted for this item")
	}

	// Tạo mới vote
	vote := &models.Vote{
		UserID:      userID,
		VotableType: votableType,
		VotableID:   votableID,
		VoteType:    models.VoteType(voteType),
	}

	// Tạo vote trong cơ sở dữ liệu
	if err := s.voteRepo.CreateVote(vote); err != nil {
		return nil, err
	}

	return vote, nil
}

func (s *voteService) GetVoteByID(id uint) (*models.Vote, error) {
	return s.voteRepo.GetVoteByID(id)
}

func (s *voteService) UpdateVote(id uint, voteType string) (*models.Vote, error) {
	// Kiểm tra valid voteType
	if voteType != string(models.VoteUp) && voteType != string(models.VoteDown) {
		return nil, errors.New("invalid vote type")
	}

	// Lấy vote hiện tại
	vote, err := s.voteRepo.GetVoteByID(id)
	if err != nil {
		return nil, err
	}

	// Cập nhật loại vote
	vote.VoteType = models.VoteType(voteType)

	// Lưu vào cơ sở dữ liệu
	if err := s.voteRepo.UpdateVote(vote); err != nil {
		return nil, err
	}

	return vote, nil
}

func (s *voteService) DeleteVote(id uint) error {
	return s.voteRepo.DeleteVote(id)
}

func (s *voteService) ListVotes() ([]models.Vote, error) {
	return s.voteRepo.ListVotes()
}

func (s *voteService) GetVoteByUserAndVotable(userID uint, votableType string, votableID uint) (*models.Vote, error) {
	return s.voteRepo.GetVoteByUserAndVotable(userID, votableType, votableID)
}

// GetVoteCount trả về số lượng upvote và downvote cho một đối tượng
func (s *voteService) GetVoteCount(votableType string, votableID uint) (int64, error) {
	return s.voteRepo.GetVoteCount(votableType, votableID)
}
