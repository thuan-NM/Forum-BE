package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"bytes"
	"encoding/json"
	"net/http"
)

type QuestionSuggestionService interface {
	FindSimilarQuestions(text string) ([]models.Question, error)
}

type questionSuggestionService struct {
	questionRepo repositories.QuestionRepository
}

func NewQuestionSuggestionService(qr repositories.QuestionRepository) QuestionSuggestionService {
	return &questionSuggestionService{questionRepo: qr}
}

func (s *questionSuggestionService) FindSimilarQuestions(text string) ([]models.Question, error) {
	embedding, err := getEmbedding(text)
	if err != nil {
		return nil, err
	}

	ids, err := searchSimilar(embedding, 3)
	if err != nil {
		return nil, err
	}

	return s.questionRepo.GetQuestionsByIDs(ids)
}

func getEmbedding(text string) ([]float32, error) {
	body := map[string]string{"text": text}
	data, _ := json.Marshal(body)

	resp, err := http.Post("http://127.0.0.1:7000/embed", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}

func searchSimilar(vector []float32, k int) ([]int, error) {
	body := map[string]interface{}{
		"vector": vector,
		"k":      k,
	}
	data, _ := json.Marshal(body)

	resp, err := http.Post("http://127.0.0.1:7000/search", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			ID       int     `json:"id"`
			Distance float64 `json:"distance"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(result.Results))
	for _, r := range result.Results {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
