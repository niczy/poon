package testutil

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CLIRunner helps run CLI commands in tests
type CLIRunner struct {
	WorkDir string
	BinPath string
}

// NewCLIRunner creates a new CLI runner
func NewCLIRunner(t *testing.T, workDir string) *CLIRunner {
	// Check if CLI binary already exists
	binPath := filepath.Join(workDir, "poon-test")
	if _, err := os.Stat(binPath); err == nil {
		// Binary exists, use it
		return &CLIRunner{
			WorkDir: workDir,
			BinPath: binPath,
		}
	}
	
	// Build CLI binary for testing from its directory
	buildCmd := exec.Command("go", "build", "-o", binPath)
	buildCmd.Dir = "../poon-cli"
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	
	return &CLIRunner{
		WorkDir: workDir,
		BinPath: binPath,
	}
}

// RunCommand executes a CLI command and returns output
func (c *CLIRunner) RunCommand(t *testing.T, args ...string) *CommandResult {
	cmd := exec.Command(c.BinPath, args...)
	cmd.Dir = c.WorkDir
	
	output, err := cmd.CombinedOutput()
	
	return &CommandResult{
		Output:   string(output),
		Error:    err,
		ExitCode: cmd.ProcessState.ExitCode(),
	}
}

// RunCommandWithServer runs a CLI command with server addresses
func (c *CLIRunner) RunCommandWithServer(t *testing.T, server *TestServer, args ...string) *CommandResult {
	fullArgs := append(args,
		"--server", server.GetGrpcAddr(),
		"--git-server", server.GetHttpURL()[7:], // Remove http://
	)
	return c.RunCommand(t, fullArgs...)
}

// CommandResult holds the result of a CLI command execution
type CommandResult struct {
	Output   string
	Error    error
	ExitCode int
}

// AssertSuccess asserts that the command succeeded
func (r *CommandResult) AssertSuccess(t *testing.T) *CommandResult {
	if r.Error != nil {
		t.Fatalf("Command failed with error: %v\nOutput: %s", r.Error, r.Output)
	}
	return r
}

// AssertError asserts that the command failed
func (r *CommandResult) AssertError(t *testing.T) *CommandResult {
	if r.Error == nil {
		t.Fatalf("Expected command to fail, but it succeeded\nOutput: %s", r.Output)
	}
	return r
}

// AssertContains asserts that output contains the given string
func (r *CommandResult) AssertContains(t *testing.T, expected string) *CommandResult {
	if !strings.Contains(r.Output, expected) {
		t.Fatalf("Expected output to contain %q, but got: %s", expected, r.Output)
	}
	return r
}

// AssertNotContains asserts that output does not contain the given string
func (r *CommandResult) AssertNotContains(t *testing.T, unexpected string) *CommandResult {
	if strings.Contains(r.Output, unexpected) {
		t.Fatalf("Expected output to not contain %q, but got: %s", unexpected, r.Output)
	}
	return r
}

// WorkspaceHelper provides utilities for workspace testing
type WorkspaceHelper struct {
	Path string
}

// NewWorkspaceHelper creates a new workspace helper
func NewWorkspaceHelper(workDir string) *WorkspaceHelper {
	return &WorkspaceHelper{Path: workDir}
}

// HasPoonDirectory checks if .poon directory exists
func (w *WorkspaceHelper) HasPoonDirectory(t *testing.T) bool {
	_, err := os.Stat(filepath.Join(w.Path, ".poon"))
	return err == nil
}

// HasGitDirectory checks if .git directory exists
func (w *WorkspaceHelper) HasGitDirectory(t *testing.T) bool {
	_, err := os.Stat(filepath.Join(w.Path, ".git"))
	return err == nil
}

// GetConfig reads the poon configuration
func (w *WorkspaceHelper) GetConfig(t *testing.T) map[string]interface{} {
	configPath := filepath.Join(w.Path, ".poon", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	
	return config
}

// GetState reads the poon state
func (w *WorkspaceHelper) GetState(t *testing.T) map[string]interface{} {
	statePath := filepath.Join(w.Path, ".poon", "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		// State file might not exist yet, return empty
		return make(map[string]interface{})
	}
	
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("Failed to parse state: %v", err)
	}
	
	return state
}

// CreateTestFile creates a test file in the workspace
func (w *WorkspaceHelper) CreateTestFile(t *testing.T, path, content string) {
	fullPath := filepath.Join(w.Path, path)
	dir := filepath.Dir(fullPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}
}

// RunGitCommand runs a git command in the workspace
func (w *WorkspaceHelper) RunGitCommand(t *testing.T, args ...string) *CommandResult {
	cmd := exec.Command("git", args...)
	cmd.Dir = w.Path
	
	output, err := cmd.CombinedOutput()
	
	return &CommandResult{
		Output:   string(output),
		Error:    err,
		ExitCode: cmd.ProcessState.ExitCode(),
	}
}