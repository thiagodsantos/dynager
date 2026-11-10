package config

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/usecase"
)

type FileRepository struct {
	path string
	mu   sync.Mutex
}

type fileData struct {
	Connections []usecase.Connection `json:"connections"`
}

func NewFileRepository(path string) (*FileRepository, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.New("invalid config path")
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0o700); err != nil {
		return nil, errors.New("could not create config directory")
	}
	return &FileRepository{path: absolute}, nil
}

func (r *FileRepository) List(ctx context.Context) ([]usecase.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return nil, err
	}
	return append([]usecase.Connection{}, connections...), nil
}

func (r *FileRepository) Get(ctx context.Context, id string) (usecase.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return usecase.Connection{}, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return usecase.Connection{}, err
	}
	for _, connection := range connections {
		if connection.ID == id {
			return connection, nil
		}
	}
	return usecase.Connection{}, &usecase.Error{Code: usecase.NotFound, Message: "Connection not found"}
}

func (r *FileRepository) Create(ctx context.Context, connection usecase.Connection) (usecase.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return usecase.Connection{}, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return usecase.Connection{}, err
	}
	for _, existing := range connections {
		if existing.ID == connection.ID {
			return usecase.Connection{}, &usecase.Error{Code: usecase.Conflict, Message: "Connection already exists"}
		}
	}
	connections = append(connections, connection)
	if err := r.writeLocked(connections); err != nil {
		return usecase.Connection{}, err
	}
	return connection, nil
}

func (r *FileRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	connections, err := r.readLocked()
	if err != nil {
		return err
	}
	for index, connection := range connections {
		if connection.ID == id {
			connections = append(connections[:index], connections[index+1:]...)
			return r.writeLocked(connections)
		}
	}
	return &usecase.Error{Code: usecase.NotFound, Message: "Connection not found"}
}

func (r *FileRepository) readLocked() ([]usecase.Connection, error) {
	file, err := os.Open(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []usecase.Connection{}, nil
	}
	if err != nil {
		return nil, errors.New("could not read connection configuration")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var data fileData
	if err := decoder.Decode(&data); err != nil {
		return nil, errors.New("connection configuration is invalid")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("connection configuration is invalid")
	}
	seen := make(map[string]struct{}, len(data.Connections))
	for _, connection := range data.Connections {
		clean, err := usecase.ValidateConnectionInput(usecase.ConnectionInput{Name: connection.Name, Region: connection.Region, Profile: connection.Profile, Endpoint: connection.Endpoint})
		if connection.ID == "" || err != nil || clean.Name != connection.Name || clean.Region != connection.Region || clean.Profile != connection.Profile || clean.Endpoint != connection.Endpoint {
			return nil, errors.New("connection configuration is invalid")
		}
		if _, exists := seen[connection.ID]; exists {
			return nil, errors.New("connection configuration is invalid")
		}
		seen[connection.ID] = struct{}{}
	}
	return data.Connections, nil
}

func (r *FileRepository) writeLocked(connections []usecase.Connection) error {
	content, err := json.MarshalIndent(fileData{Connections: connections}, "", "  ")
	if err != nil {
		return errors.New("could not encode connection configuration")
	}
	content = append(content, '\n')
	directory := filepath.Dir(r.path)
	temporary, err := os.CreateTemp(directory, ".connections-*.tmp")
	if err != nil {
		return errors.New("could not write connection configuration")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return errors.New("could not secure connection configuration")
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return errors.New("could not write connection configuration")
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return errors.New("could not write connection configuration")
	}
	if err := temporary.Close(); err != nil {
		return errors.New("could not write connection configuration")
	}
	if err := os.Rename(temporaryPath, r.path); err != nil {
		return errors.New("could not replace connection configuration")
	}
	if directoryFile, err := os.Open(directory); err == nil {
		_ = directoryFile.Sync()
		_ = directoryFile.Close()
	}
	return nil
}
