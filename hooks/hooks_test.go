package hooks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kupospelov/feeds-to-instapaper/config"
	"github.com/mmcdole/gofeed"
)

func TestSuccess(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "output")
	conf := config.Hooks{
		NewArticle: []config.Hook{
			{
				Spawn: []string{"sh", "-c", fmt.Sprintf("echo -n {{.Feed.Title}}, {{.Feed.Link}}, {{.Title}} > %s", file)},
			},
		},
	}
	hooks, err := New(conf)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	item := &gofeed.Item{Link: "http://example.com/1", Title: "Article 1"}
	feed := &gofeed.Feed{
		Link:  "http://example.com/",
		Title: "Feed Title",
		Items: []*gofeed.Item{item},
	}
	hooks.NewArticle(feed, item)

	content, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}
	if got, want := string(content), "Feed Title, http://example.com, Article 1"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNoConfiguredCommand(t *testing.T) {
	confs := []config.Hooks{
		{
			NewArticle: []config.Hook{
				{}, // Empty hook
			},
		},
		{
			NewArticle: []config.Hook{
				{
					Spawn: []string{}, // Empty spawn command
				},
			},
		},
	}

	for _, conf := range confs {
		_, err := New(conf)
		if !errors.Is(err, ErrNoSpawnCommand) {
			t.Errorf("expected %v, got %v", ErrNoSpawnCommand, err)
		}
	}
}

func TestInvalidTemplate(t *testing.T) {
	confs := []config.Hooks{
		{
			NewArticle: []config.Hook{
				{
					Spawn: []string{
						"echo",
						"{{.Title",
					},
				},
			},
		},
	}

	for _, conf := range confs {
		_, err := New(conf)
		if !errors.Is(err, ErrInvalidTemplate) {
			t.Errorf("expected %v, got %v", ErrInvalidTemplate, err)
		}
	}
}
