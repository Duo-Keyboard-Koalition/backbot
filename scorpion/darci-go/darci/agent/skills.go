package agent

import (
	"context"
	"sync"

	"darci-go/internal/adk"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string
	Description string
	Tool        adk.Tool
	Enabled     bool
}

// SkillsLoader manages skill loading.
type SkillsLoader struct {
	mu     sync.RWMutex
	skills map[string]*Skill
}

// NewSkillsLoader creates a new skills loader.
func NewSkillsLoader() *SkillsLoader {
	return &SkillsLoader{
		skills: make(map[string]*Skill),
	}
}

// Load loads skills from the skills directory.
func (s *SkillsLoader) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In a full implementation, this would:
	// 1. Scan the skills directory
	// 2. Parse skill definitions (SKILL.md files)
	// 3. Register skill handlers
	// For now, this is a placeholder

	return nil
}

// Register registers a skill.
func (s *SkillsLoader) Register(skill *Skill) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if skill.Name == "" {
		return ErrSkillNameRequired
	}

	s.skills[skill.Name] = skill
	return nil
}

// Get retrieves a skill by name.
func (s *SkillsLoader) Get(name string) (*Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	skill, ok := s.skills[name]
	return skill, ok
}

// List returns all loaded skills.
func (s *SkillsLoader) List() []*Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	skills := make([]*Skill, 0, len(s.skills))
	for _, skill := range s.skills {
		skills = append(skills, skill)
	}
	return skills
}

// Enable enables a skill.
func (s *SkillsLoader) Enable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	skill, ok := s.skills[name]
	if !ok {
		return ErrSkillNotFound
	}

	skill.Enabled = true
	return nil
}

// Disable disables a skill.
func (s *SkillsLoader) Disable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	skill, ok := s.skills[name]
	if !ok {
		return ErrSkillNotFound
	}

	skill.Enabled = false
	return nil
}

// Unregister removes a skill.
func (s *SkillsLoader) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.skills, name)
}

// Execute executes a skill.
func (s *SkillsLoader) Execute(ctx context.Context, name string, args map[string]string) (string, error) {
	s.mu.RLock()
	skill, ok := s.skills[name]
	s.mu.RUnlock()

	if !ok {
		return "", ErrSkillNotFound
	}

	if !skill.Enabled {
		return "", ErrSkillDisabled
	}

	return skill.Tool.Run(ctx, args)
}

// Errors
var (
	ErrSkillNameRequired = &skillError{"skill name is required"}
	ErrSkillNotFound     = &skillError{"skill not found"}
	ErrSkillDisabled     = &skillError{"skill is disabled"}
)

type skillError struct {
	message string
}

func (e *skillError) Error() string {
	return e.message
}
