package service

import (
	"sync"
)

// RatingScore is an interface to store laptop ratings.
type RatingStore interface {
	// Add adds a new laptop score to the store and returns its rating.
	Add(laptopID string, score float64) (*Rating, error)
}

// A Rating contains the rating information of a laptop.
type Rating struct {
	Count uint32
	Sum   float64
}

// InMemoryRatingStore stores laptop ratings in memory
type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore instance.
func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

// Add adds a new laptop to the store and returns its rating
func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopID] = rating
	return rating, nil
}
