package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/IkehAkinyemi/pcbook/pb"
	"github.com/jinzhu/copier"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in the store.
var ErrAlreadyExists = errors.New("records already exists")
var ErrNotFound = errors.New("record not found")

// A LaptopStore is an interface to store laptop.
type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by ID
	Find(id string) (*pb.Laptop, error)
	// Search searches for laptops with filter, returns one by one via the found function
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// A InMemoryLaptopStore stores laptop in memory.
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore.
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save saves the laptop to the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	copy, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[copy.Id] = copy

	return nil
}

// Find finds a laptop by ID
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop, ok := store.data[id]
	if !ok {
		return nil, ErrNotFound
	}

	return deepCopy(laptop)
}

// Search returns laptops that match the search criteria in filter.
func (store *InMemoryLaptopStore) Search(
	ctx context.Context,
	filter *pb.Filter,
	found func(laptop *pb.Laptop) error,
) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Printf("context is cancelled")
			return errors.New("context is cancelled")
		}
		if isQualified(filter, laptop) {
			copy, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(copy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	val := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return val
	case pb.Memory_BYTE:
		return val << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return val << 13 // 13 = 2^10 * 2^3= 2^13
	case pb.Memory_MEGABYTE:
		return val << 23
	case pb.Memory_GIGABYTE:
		return val << 33
	case pb.Memory_TERABYTE:
		return val << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	tmp := &pb.Laptop{}
	err := copier.Copy(tmp, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return tmp, nil
}
