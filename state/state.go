package state

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type State struct {
	Path           string
	ProcessedItems sync.Map
	NewItems       []string
}

func EmptyWithPath(path string) *State {
	return &State{
		Path:           path,
		ProcessedItems: sync.Map{},
		NewItems:       make([]string, 0),
	}
}

func New() *State {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		stateDir = filepath.Join(homeDir, ".local", "state")
	}

	return EmptyWithPath(filepath.Join(stateDir, "feeds-to-instapaper", "added"))
}

func Load() (*State, error) {
	state := New()

	if _, err := os.Stat(state.Path); os.IsNotExist(err) {
		return state, nil
	}

	file, err := os.Open(state.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open state file %s: %w", state.Path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		state.MarkProcessed(scanner.Text())
	}

	return state, nil
}

func (s *State) Save() error {
	if len(s.NewItems) < 1 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(s.Path), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	file, err := os.OpenFile(s.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create state file: %w", err)
	}
	defer file.Close()

	for _, item := range s.NewItems {
		file.WriteString(item)
		file.WriteString("\n")
	}

	return nil
}

func (s *State) Append(item string) {
	s.NewItems = append(s.NewItems, item)
}

// MarkProcessed returns true if the item has not been marked before; otherwise, returns false.
func (s *State) MarkProcessed(item string) bool {
	_, loaded := s.ProcessedItems.LoadOrStore(item, struct{}{})
	return !loaded
}
