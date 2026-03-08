// Package utils provides utility functions for darci-go.
package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// EnsureDir ensures the directory exists, creating it if necessary.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetDataPath returns the ~/.darci-go data directory.
func GetDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, ".darci-go")
	return path, EnsureDir(path)
}

// GetWorkspacePath resolves and returns the workspace path.
// Defaults to ~/.darci-go/workspace if not specified.
func GetWorkspacePath(workspace string) (string, error) {
	if workspace == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		workspace = filepath.Join(homeDir, ".darci-go", "workspace")
	}

	// Expand tilde
	if strings.HasPrefix(workspace, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		workspace = filepath.Join(homeDir, workspace[1:])
	}

	return workspace, EnsureDir(workspace)
}

// Timestamp returns the current ISO timestamp.
func Timestamp() string {
	return time.Now().Format(time.RFC3339)
}

// unsafeChars matches characters unsafe for filenames.
var unsafeChars = regexp.MustCompile(`[<>:"/\\|?*]`)

// SafeFilename replaces unsafe path characters with underscores.
func SafeFilename(name string) string {
	return strings.TrimSpace(unsafeChars.ReplaceAllString(name, "_"))
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

// ReadFile reads a file and returns its content.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes content to a file.
func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	content, err := ReadFile(src)
	if err != nil {
		return err
	}
	return WriteFile(dst, content)
}

// JoinPath joins path elements.
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// AbsPath returns the absolute path.
func AbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// BaseName returns the base name of a path.
func BaseName(path string) string {
	return filepath.Base(path)
}

// DirName returns the directory name of a path.
func DirName(path string) string {
	return filepath.Dir(path)
}

// Ext returns the file extension.
func Ext(path string) string {
	return filepath.Ext(path)
}

// TrimExt returns the path without extension.
func TrimExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
