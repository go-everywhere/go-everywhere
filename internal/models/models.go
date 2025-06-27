package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Model struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	OriginalImage string `json:"original_image,omitempty"`
}

type ModelStore struct {
	mu      sync.RWMutex
	models  []Model
	dataFile string
}

func NewModelStore(dataDir string) (*ModelStore, error) {
	dataFile := filepath.Join(dataDir, "models.json")
	store := &ModelStore{
		dataFile: dataFile,
		models:   []Model{},
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// Load existing models
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

func (s *ModelStore) load() error {
	data, err := ioutil.ReadFile(s.dataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.models)
}

func (s *ModelStore) save() error {
	data, err := json.MarshalIndent(s.models, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.dataFile, data, 0644)
}

func (s *ModelStore) AddModel(id string, originalImage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	model := Model{
		ID:            id,
		CreatedAt:     time.Now(),
		OriginalImage: originalImage,
	}

	s.models = append(s.models, model)
	return s.save()
}

func (s *ModelStore) GetAll() []Model {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]Model, len(s.models))
	copy(result, s.models)

	// Sort by creation time (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].CreatedAt.After(result[i].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func (s *ModelStore) GetByID(id string) (*Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, model := range s.models {
		if model.ID == id {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", id)
}

func (s *ModelStore) DeleteModel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newModels := []Model{}
	found := false

	for _, model := range s.models {
		if model.ID != id {
			newModels = append(newModels, model)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("model not found: %s", id)
	}

	s.models = newModels
	return s.save()
}