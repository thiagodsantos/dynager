package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/usecase"
)

func TestFileRepositoryPersistsOnlyNonSecretConnectionFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "connections.json")
	repository, err := NewFileRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	connection := usecase.Connection{ID: "id-1", Name: "local", Region: "us-east-1", Profile: "dev", Endpoint: "http://localhost:8000"}
	if _, err := repository.Create(context.Background(), connection); err != nil {
		t.Fatal(err)
	}
	loaded, err := repository.Get(context.Background(), "id-1")
	if err != nil || loaded != connection {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"secret", "accessKey", "password"} {
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("configuration contains a secret field: %s", forbidden)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
	if err := repository.Delete(context.Background(), "id-1"); err != nil {
		t.Fatal(err)
	}
	connections, err := repository.List(context.Background())
	if err != nil || len(connections) != 0 {
		t.Fatalf("connections=%v err=%v", connections, err)
	}
}

func TestFileRepositoryRejectsUnknownSecretFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	if err := os.WriteFile(path, []byte(`{"connections":[{"id":"x","name":"n","region":"us-east-1","secret":"do-not-load"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	repository, err := NewFileRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.List(context.Background()); err == nil {
		t.Fatal("expected unknown field to be rejected")
	}
}
