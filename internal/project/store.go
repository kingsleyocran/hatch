package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Dir     string    `yaml:"dir"`
	Domain  string    `yaml:"domain"`
	Port    int       `yaml:"port"`
	HTTPS   bool      `yaml:"https"`
	Created time.Time `yaml:"created"`
}

type storeData struct {
	Projects []Project `yaml:"projects"`
}

type Store struct {
	mu       sync.RWMutex
	projects []Project
	path     string
}

func Load(path string) (*Store, error) {
	s := &Store{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	var sd storeData
	if err := yaml.Unmarshal(data, &sd); err != nil {
		return nil, err
	}
	s.projects = sd.Projects

	return s, nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	sd := storeData{Projects: s.projects}
	data, err := yaml.Marshal(&sd)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Add(p Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.projects {
		if existing.Domain == p.Domain {
			return fmt.Errorf("domain %q already exists", p.Domain)
		}
	}

	s.projects = append(s.projects, p)
	return nil
}

func (s *Store) Remove(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.projects {
		if p.Domain == domain {
			s.projects = append(s.projects[:i], s.projects[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("domain %q not found", domain)
}

func (s *Store) List() []Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Project, len(s.projects))
	copy(out, s.projects)
	return out
}

func (s *Store) FindByDomain(domain string) (*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.projects {
		if p.Domain == domain {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("domain %q not found", domain)
}

func (s *Store) FindByPort(port int) []Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []Project
	for _, p := range s.projects {
		if p.Port == port {
			matches = append(matches, p)
		}
	}
	return matches
}
