package adk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Tool is a self-contained function the model can call.
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input map[string]string) (string, error)
}

// ToolRegistry stores all available tools by name.
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: map[string]Tool{}}
}

func (r *ToolRegistry) Register(t Tool) {
	if t == nil {
		return
	}
	r.tools[t.Name()] = t
}

func (r *ToolRegistry) Execute(ctx context.Context, call ToolCall) ToolResult {
	tool, ok := r.tools[call.Name]
	if !ok {
		return ToolResult{Name: call.Name, Output: "unknown tool", IsError: true}
	}
	out, err := tool.Run(ctx, call.Input)
	if err != nil {
		return ToolResult{Name: call.Name, Output: err.Error(), IsError: true}
	}
	return ToolResult{Name: call.Name, Output: out, IsError: false}
}

func (r *ToolRegistry) Names() []string {
	names := make([]string, 0, len(r.tools))
	for k := range r.tools {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

type TimeNowTool struct{}

func (TimeNowTool) Name() string { return "time_now" }
func (TimeNowTool) Description() string {
	return "Return current UTC time in RFC3339 format"
}
func (TimeNowTool) Run(_ context.Context, _ map[string]string) (string, error) {
	return time.Now().UTC().Format(time.RFC3339), nil
}

type ListDirTool struct {
	BaseDir string
}

func (ListDirTool) Name() string { return "list_dir" }
func (ListDirTool) Description() string {
	return "List files in a directory under the configured workspace"
}
func (t ListDirTool) Run(_ context.Context, input map[string]string) (string, error) {
	rel := strings.TrimSpace(input["path"])
	if rel == "" {
		rel = "."
	}
	full, err := safePath(t.BaseDir, rel)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		return "", err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() {
			n += "/"
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return strings.Join(out, "\n"), nil
}

type ReadFileTool struct {
	BaseDir string
}

func (ReadFileTool) Name() string { return "read_file" }
func (ReadFileTool) Description() string {
	return "Read a UTF-8 text file under the configured workspace"
}
func (t ReadFileTool) Run(_ context.Context, input map[string]string) (string, error) {
	rel := strings.TrimSpace(input["path"])
	if rel == "" {
		return "", fmt.Errorf("path is required")
	}
	full, err := safePath(t.BaseDir, rel)
	if err != nil {
		return "", err
	}
	buf, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func safePath(baseDir, rel string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(baseAbs, rel)
	full, err := filepath.Abs(resolved)
	if err != nil {
		return "", err
	}
	if full != baseAbs && !strings.HasPrefix(full, baseAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace")
	}
	return full, nil
}
