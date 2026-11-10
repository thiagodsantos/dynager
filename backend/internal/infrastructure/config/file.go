package config

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

type FileRepository struct {
	path string
	mu   sync.Mutex
}

type fileData struct {
	Connections []domain.Connection `json:"connections"`
}

func NewFileRepository(path string) (*FileRepository, error) {
	if path == "" {
		return nil, errConfigPathRequired
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, errInvalidConfigPath
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0o700); err != nil {
		return nil, errCreateConfigDirectory
	}
	return &FileRepository{path: absolute}, nil
}

func (r *FileRepository) List(ctx context.Context) ([]domain.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return nil, err
	}
	return append([]domain.Connection{}, connections...), nil
}

func (r *FileRepository) Get(ctx context.Context, id string) (domain.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return domain.Connection{}, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return domain.Connection{}, err
	}
	for _, connection := range connections {
		if connection.ID == id {
			return connection, nil
		}
	}
	return domain.Connection{}, &domain.Error{Code: domain.NotFound, Message: "Connection not found"}
}

func (r *FileRepository) Create(ctx context.Context, connection domain.Connection) (domain.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return domain.Connection{}, err
	}
	connections, err := r.readLocked()
	if err != nil {
		return domain.Connection{}, err
	}
	for _, existing := range connections {
		if existing.ID == connection.ID {
			return domain.Connection{}, &domain.Error{Code: domain.Conflict, Message: "Connection already exists"}
		}
	}
	connections = append(connections, connection)
	if err := r.writeLocked(connections); err != nil {
		return domain.Connection{}, err
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
	return &domain.Error{Code: domain.NotFound, Message: "Connection not found"}
}

func (r *FileRepository) readLocked() ([]domain.Connection, error) {
	file, err := os.Open(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []domain.Connection{}, nil
	}
	if err != nil {
		return nil, errReadConnectionConfiguration
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var data fileData
	if err := decoder.Decode(&data); err != nil {
		return nil, errInvalidConnectionConfiguration
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errInvalidConnectionConfiguration
	}
	seen := make(map[string]struct{}, len(data.Connections))
	for _, connection := range data.Connections {
		clean, err := domain.ValidateConnectionInput(domain.ConnectionInput{Name: connection.Name, Region: connection.Region, Profile: connection.Profile, Endpoint: connection.Endpoint})
		if connection.ID == "" || err != nil || clean.Name != connection.Name || clean.Region != connection.Region || clean.Profile != connection.Profile || clean.Endpoint != connection.Endpoint {
			return nil, errInvalidConnectionConfiguration
		}
		if _, exists := seen[connection.ID]; exists {
			return nil, errInvalidConnectionConfiguration
		}
		seen[connection.ID] = struct{}{}
	}
	return data.Connections, nil
}

func (r *FileRepository) writeLocked(connections []domain.Connection) error {
	content, err := json.MarshalIndent(fileData{Connections: connections}, "", "  ")
	if err != nil {
		return errEncodeConnectionConfiguration
	}
	content = append(content, '\n')
	directory := filepath.Dir(r.path)
	temporary, err := os.CreateTemp(directory, ".connections-*.tmp")
	if err != nil {
		return errWriteConnectionConfiguration
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return errSecureConnectionConfiguration
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return errWriteConnectionConfiguration
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return errWriteConnectionConfiguration
	}
	if err := temporary.Close(); err != nil {
		return errWriteConnectionConfiguration
	}
	if err := os.Rename(temporaryPath, r.path); err != nil {
		return errReplaceConnectionConfiguration
	}
	if directoryFile, err := os.Open(directory); err == nil {
		_ = directoryFile.Sync()
		_ = directoryFile.Close()
	}
	return nil
}
