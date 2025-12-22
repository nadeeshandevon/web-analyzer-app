package repositorysql

import (
	"errors"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/repository"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/google/uuid"
)

type webAnalyzerRepo struct {
	log     *logger.Logger
	storage map[string]model.WebAnalyzer
}

func NewWebAnalyzerRepo(logger *logger.Logger) repository.WebAnalyzerRepository {
	return &webAnalyzerRepo{
		log:     logger,
		storage: make(map[string]model.WebAnalyzer),
	}
}

func (r *webAnalyzerRepo) Save(webAnalyzer model.WebAnalyzer) (string, error) {
	id := generateID()
	webAnalyzer.ID = id
	r.storage[id] = webAnalyzer
	return id, nil
}

func (r *webAnalyzerRepo) GetById(id string) (*model.WebAnalyzer, error) {
	if val, ok := r.storage[id]; ok {
		return &val, nil
	}
	return nil, nil
}

func (r *webAnalyzerRepo) Update(webAnalyzer model.WebAnalyzer) (string, error) {
	if _, ok := r.storage[webAnalyzer.ID]; !ok {
		return "", errors.New("record not found")
	}

	r.storage[webAnalyzer.ID] = webAnalyzer
	return webAnalyzer.ID, nil
}

func generateID() string {
	id := uuid.New().String()
	return id
}
