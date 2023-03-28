package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

// A ImageStore is an interface to store image files.
type ImageStore interface {
	Save(laptopID, imageType string, imageData bytes.Buffer) (string, error)
}

// A DiskImageStore stores images to disk and its info on memory.
type DiskImageStore struct {
	mutex sync.RWMutex
	imageFolder string
	images map[string]*ImageInfo
}

// A ImageInfo stores information about laptop image.
type ImageInfo struct {
	LaptopID string
	Type string
	Path string
}

// NewDiskImageStore defines and return an instance of DiskImageStore.
func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images: make(map[string]*ImageInfo),
	}
}

// Save saves a new laptop image to the store.
func (store *DiskImageStore) Save(
	laptopID,
	imageType string,
	imageData bytes.Buffer,
) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}

	imagePath := fmt.Sprintf("%s/%s.%s", store.imageFolder, imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type: imageType,
		Path: imagePath,
	}

	return imageID.String(), nil
}