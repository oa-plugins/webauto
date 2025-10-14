package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// NodeCommand represents a command to execute via Node.js
type NodeCommand struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

// NodeResponse represents a response from Node.js
type NodeResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// NodeExecutor handles Node.js subprocess execution
type NodeExecutor struct {
	nodePath string
	timeout  time.Duration
}

// NewNodeExecutor creates a new NodeExecutor
func NewNodeExecutor(nodePath string, timeout time.Duration) *NodeExecutor {
	return &NodeExecutor{
		nodePath: nodePath,
		timeout:  timeout,
	}
}

// Execute runs a Node.js command with the given script
func (ne *NodeExecutor) Execute(ctx context.Context, script string) (NodeResponse, error) {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, ne.timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(timeoutCtx, ne.nodePath, "-e", script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	if err != nil {
		return NodeResponse{
			Success: false,
			Error:   fmt.Sprintf("node execution failed: %v, stderr: %s", err, stderr.String()),
		}, err
	}

	// Parse response
	var response NodeResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return NodeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to parse node response: %v", err),
		}, err
	}

	return response, nil
}

// ExecuteJSON runs a Node.js command with JSON-encoded parameters
func (ne *NodeExecutor) ExecuteJSON(ctx context.Context, command NodeCommand) (NodeResponse, error) {
	// Encode command as JSON
	cmdJSON, err := json.Marshal(command)
	if err != nil {
		return NodeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to encode command: %v", err),
		}, err
	}

	// Build Node.js script
	script := fmt.Sprintf(`
		const cmd = %s;
		// Execute command and output result
		console.log(JSON.stringify({ success: true, data: cmd }));
	`, string(cmdJSON))

	return ne.Execute(ctx, script)
}

// CheckNode verifies that Node.js is available
func CheckNode(nodePath string) error {
	cmd := exec.Command(nodePath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Node.js not found at %s: %w", nodePath, err)
	}
	return nil
}
