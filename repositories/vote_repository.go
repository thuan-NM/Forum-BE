package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type VoteRepository interface {
	CreateVote(vote *models.Vote) error
	GetVoteByID(id uint) (*models.Vote, error)
	UpdateVote(vote *models.Vote) error
	DeleteVote(id uint) error
	ListVotes() ([]models.Vote, error)
	GetVoteByUserAndVotable(userID uint, votableType string, votableID uint) (*models.Vote, error)
	GetVoteCount(votableType string, votableID uint) (int64, error)
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) CreateVote(vote *models.Vote) error {
	return r.db.Create(vote).Error
}

func (r *voteRepository) GetVoteByID(id uint) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.First(&vote, id).Error
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *voteRepository) UpdateVote(vote *models.Vote) error {
	return r.db.Save(vote).Error
}

func (r *voteRepository) DeleteVote(id uint) error {
	return r.db.Delete(&models.Vote{}, id).Error
}

func (r *voteRepository) ListVotes() ([]models.Vote, error) {
	var votes []models.Vote
	err := r.db.Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}

func (r *voteRepository) GetVoteByUserAndVotable(userID uint, votableType string, votableID uint) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.Where("user_id = ? AND votable_type = ? AND votable_id = ?", userID, votableType, votableID).First(&vote).Error
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *voteRepository) GetVoteCount(votableType string, votableID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Vote{}).
		Where("votable_type = ? AND votable_id = ?", votableType, votableID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
