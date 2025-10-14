package response

import (
	"encoding/json"
	"os"
	"time"
)

// StandardResponse represents the standard JSON response structure for all commands
type StandardResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Error    *ErrorInfo  `json:"error"`
	Metadata Metadata    `json:"metadata"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code               string      `json:"code"`
	Message            string      `json:"message"`
	Details            interface{} `json:"details,omitempty"`
	RecoverySuggestion string      `json:"recovery_suggestion,omitempty"`
}

// Metadata contains plugin metadata
type Metadata struct {
	Plugin          string `json:"plugin"`
	Version         string `json:"version"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}

// Success creates a success response
func Success(data interface{}, startTime time.Time) *StandardResponse {
	return &StandardResponse{
		Success: true,
		Data:    data,
		Error:   nil,
		Metadata: Metadata{
			Plugin:          "webauto",
			Version:         "1.0.0",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		},
	}
}

// Error creates an error response
func Error(code, message, recovery string, details interface{}, startTime time.Time) *StandardResponse {
	return &StandardResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorInfo{
			Code:               code,
			Message:            message,
			Details:            details,
			RecoverySuggestion: recovery,
		},
		Metadata: Metadata{
			Plugin:          "webauto",
			Version:         "1.0.0",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		},
	}
}

// Print outputs the response as formatted JSON to stdout
func (r *StandardResponse) Print() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(r)
}
