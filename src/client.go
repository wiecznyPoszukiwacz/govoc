// Package main contains the JSON-RPC client logic
// This file handles all network communication, request creation, and response parsing
package main

// Import section - we need these libraries for JSON-RPC functionality
import (
	"bytes"         // For creating byte buffers (needed for HTTP request body)
	"encoding/json" // For converting Go structs to/from JSON format
	"fmt"           // For formatted string printing and error creation
	"io"            // For reading data streams (like HTTP response body)
	"net/http"      // For making HTTP requests to the server
	"strings"       // For string manipulation (like trimming whitespace)

	tea "github.com/charmbracelet/bubbletea" // For creating tea commands
	"gopkg.in/yaml.v3"                       // YAML parsing for parameter conversion
)

// JSONRPCRequest represents a JSON-RPC 2.0 request message
// This struct defines the structure that will be sent to the server
// The `json:"..."` tags tell Go how to name fields when converting to/from JSON
type JSONRPCRequest struct {
	// JSONRPC field must always be "2.0" for JSON-RPC 2.0 protocol
	JSONRPC string `json:"jsonrpc"`

	// Method is the name of the remote procedure we want to call
	Method string `json:"method"`

	// Params contains the arguments for the method call
	// interface{} means "any type" - could be object, array, string, etc.
	// omitempty means this field won't appear in JSON if it's empty/nil
	Params interface{} `json:"params,omitempty"`

	// ID is a unique identifier for this request
	// The server will include this same ID in its response
	ID int `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response message
// This is what we expect to receive back from the server
type JSONRPCResponse struct {
	// JSONRPC should be "2.0" in the response
	JSONRPC string `json:"jsonrpc"`

	// Result contains the successful result (if no error occurred)
	// interface{} means the result can be any JSON type
	Result interface{} `json:"result,omitempty"`

	// Error contains error information (if something went wrong)
	// *JSONRPCError means "pointer to JSONRPCError" - can be nil if no error
	Error *JSONRPCError `json:"error,omitempty"`

	// ID should match the ID from our request
	ID int `json:"id"`
}

// JSONRPCError represents an error in a JSON-RPC response
// This follows the JSON-RPC 2.0 error object specification
type JSONRPCError struct {
	// Code is a number indicating the error type
	// Standard codes: -32700 (Parse error), -32600 (Invalid Request), etc.
	Code int `json:"code"`

	// Message is a human-readable description of the error
	Message string `json:"message"`

	// Data provides additional error information (optional)
	// Can be any JSON type - string, object, array, etc.
	Data interface{} `json:"data,omitempty"`
}

// ResponseMsg is our custom message type for communicating between goroutines
// This is how we send results back to the UI after the network request completes
// In bubbletea, all communication happens through messages
type ResponseMsg struct {
	// Response contains the formatted response text to display to user
	Response string

	// Err contains any error that occurred during the request
	// In Go, errors are just values (not exceptions like in other languages)
	Err error

	// History data for saving request/response to disk
	RequestData  string // The raw JSON-RPC request that was sent
	ResponseData string // The raw JSON-RPC response that was received
	StatusCode   int    // HTTP status code
	Success      bool   // Whether the JSON-RPC call succeeded (no error field)
	ErrorMessage string // Error message if the call failed
}

// yamlToJsonForClient converts YAML string to JSON for API communication
// This is the client-side version that handles parameter conversion
func yamlToJsonForClient(yamlStr string) (string, error) {
	// Handle empty or whitespace-only input
	yamlStr = strings.TrimSpace(yamlStr)
	if yamlStr == "" {
		return "{}", nil // Return empty JSON object for empty YAML
	}
	
	// Parse YAML into a generic interface
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &yamlData); err != nil {
		return "", fmt.Errorf("invalid YAML parameters: %v", err)
	}
	
	// Convert to JSON
	jsonBytes, err := json.Marshal(yamlData)
	if err != nil {
		return "", fmt.Errorf("failed to convert YAML to JSON: %v", err)
	}
	
	return string(jsonBytes), nil
}

// jsonToYamlForClient converts JSON string to YAML string for display
func jsonToYamlForClient(jsonStr string) string {
	// Handle empty or whitespace-only input
	if strings.TrimSpace(jsonStr) == "" {
		return ""
	}
	
	// Parse JSON into a generic interface
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		// If JSON parsing fails, return original string
		return jsonStr
	}
	
	// Convert to YAML
	yamlBytes, err := yaml.Marshal(jsonData)
	if err != nil {
		// If YAML marshaling fails, return original string
		return jsonStr
	}
	
	return strings.TrimSpace(string(yamlBytes))
}

// MakeRequest creates and sends a JSON-RPC request to the server
// It takes our Model (containing user input) and returns a tea.Cmd
// tea.Cmd represents an asynchronous operation that bubbletea will execute
func MakeRequest(m Model) tea.Cmd {
	// tea.Cmd is actually a function type: func() tea.Msg
	// We return a function that will be called in a separate goroutine
	// This prevents the UI from freezing while we wait for the network request
	return tea.Cmd(func() tea.Msg {
		// Step 1: Parse the JSON parameters that the user typed
		// We declare params as interface{} so it can hold any JSON value
		var params interface{}

		// Convert YAML parameters to JSON and then parse them
		// We only try to parse if there's actually some text
		if strings.TrimSpace(m.ParamsText) != "" {
			// First convert YAML to JSON
			jsonStr, err := yamlToJsonForClient(m.ParamsText)
			if err != nil {
				// If YAML conversion fails, return an error message immediately
				return ResponseMsg{
					Response:     "",                                            // No response text
					Err:          fmt.Errorf("invalid YAML parameters: %v", err), // %v formats any value
					RequestData:  "",                                            // No request was made
					ResponseData: "",                                            // No response received
					StatusCode:   0,                                             // No HTTP status
					Success:      false,                                         // Failed validation
					ErrorMessage: "Invalid YAML parameters",                    // Error description
				}
			}
			
			// Now parse the converted JSON into Go values
			// json.Unmarshal converts JSON text into Go values
			// []byte(jsonStr) converts string to byte slice (required by json.Unmarshal)
			// &params means "write the result into the params variable"
			if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
				// If JSON parsing fails after YAML conversion, return an error
				return ResponseMsg{
					Response:     "",                                            // No response text
					Err:          fmt.Errorf("failed to parse converted parameters: %v", err), // %v formats any value
					RequestData:  "",                                            // No request was made
					ResponseData: "",                                            // No response received
					StatusCode:   0,                                             // No HTTP status
					Success:      false,                                         // Failed validation
					ErrorMessage: "Invalid parameter format",                   // Error description
				}
			}
		}
		// If ParamsText is empty, params stays nil (which is fine for JSON-RPC)

		// Step 2: Create the JSON-RPC request object
		// We use struct literal syntax to create and initialize the struct
		request := JSONRPCRequest{
			JSONRPC: "2.0",          // Always "2.0" for JSON-RPC 2.0
			Method:  m.SelectedMethod, // Get currently selected method (from list or custom)
			Params:  params,         // The parsed parameters (could be nil)
			ID:      m.RequestID,    // Unique ID for this request
		}

		// Step 3: Convert the request struct to JSON bytes
		// json.Marshal converts Go values to JSON format
		// It returns []byte (byte slice) and error
		requestBody, err := json.Marshal(request)
		if err != nil {
			// This should rarely happen unless there's a bug in our code
			return ResponseMsg{
				Response:     "",
				Err:          fmt.Errorf("failed to marshal request: %v", err),
				RequestData:  "",
				ResponseData: "",
				StatusCode:   0,
				Success:      false,
				ErrorMessage: "Failed to marshal request",
			}
		}

		// Step 4: Send HTTP POST request to the server
		// http.Post(url, contentType, body) sends a POST request
		// bytes.NewBuffer(requestBody) creates an io.Reader from our byte slice
		// "application/json" tells the server we're sending JSON data
		resp, err := http.Post(m.ServerURL, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			// Network errors: server unreachable, DNS failure, timeout, etc.
			return ResponseMsg{
				Response:     "",
				Err:          fmt.Errorf("network error: %v", err),
				RequestData:  string(requestBody), // We have the request data
				ResponseData: "",                  // No response received
				StatusCode:   0,                   // No status code
				Success:      false,               // Network error
				ErrorMessage: "Network error",     // Error description
			}
		}

		// defer means "execute this when the function returns"
		// Always close response body to prevent memory leaks
		// This is important in Go - you must close resources when done
		defer resp.Body.Close()

		// Step 5: Read the entire response body
		// io.ReadAll reads all data from an io.Reader until EOF
		// resp.Body is an io.Reader containing the server's response
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			// Error reading the response (connection interrupted, etc.)
			return ResponseMsg{
				Response:     "",
				Err:          fmt.Errorf("failed to read response: %v", err),
				RequestData:  string(requestBody), // We have the request data
				ResponseData: "",                  // Failed to read response
				StatusCode:   resp.StatusCode,     // We have status code
				Success:      false,               // Read error
				ErrorMessage: "Failed to read response", // Error description
			}
		}

		// Step 6: Check HTTP status code
		// http.StatusOK is constant 200 (success)
		// JSON-RPC can return application errors even with HTTP 200
		// But HTTP errors (404, 500, etc.) indicate server/network problems
		if resp.StatusCode != http.StatusOK {
			return ResponseMsg{
				Response:     "",
				// Include both status code and response body in error message
				Err:          fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody)),
				RequestData:  string(requestBody),  // We have the request data
				ResponseData: string(responseBody), // We have the response data
				StatusCode:   resp.StatusCode,      // Non-OK status code
				Success:      false,                // HTTP error
				ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode), // Error description
			}
		}

		// Step 7: Parse the JSON-RPC response
		// Declare a variable to hold the parsed response
		var jsonRPCResp JSONRPCResponse

		// Try to parse the response as JSON-RPC format
		if err := json.Unmarshal(responseBody, &jsonRPCResp); err != nil {
			// If parsing fails, maybe the server returned non-JSON data
			// Show the raw response to the user for debugging
			return ResponseMsg{
				Response:     string(responseBody), // Convert byte slice to string
				Err:          nil,                  // Not really an error, just unexpected format
				RequestData:  string(requestBody),  // We have the request data
				ResponseData: string(responseBody), // Non-JSON response
				StatusCode:   resp.StatusCode,      // HTTP status code
				Success:      false,                // Not valid JSON-RPC
				ErrorMessage: "Invalid JSON-RPC response", // Error description
			}
		}

		// Step 8: Format the response for display to the user
		// We need to handle two cases: success (result) or error
		var formattedResponse string

		// Check if the JSON-RPC response contains an error
		// jsonRPCResp.Error is a pointer, so we check if it's not nil
		if jsonRPCResp.Error != nil {
			// Format error message with emoji, code, message, and additional data
			formattedResponse = fmt.Sprintf("❌ JSON-RPC Error [%d]: %s\nData: %v",
				jsonRPCResp.Error.Code,    // Error code (like -32600)
				jsonRPCResp.Error.Message, // Human-readable error message
				jsonRPCResp.Error.Data)    // Additional error data (could be nil)
		} else {
			// Success case - format the result for display as YAML
			// First convert to JSON, then to YAML for better readability
			resultJSON, err := json.Marshal(jsonRPCResp.Result)
			if err != nil {
				// If JSON marshaling fails, show result without formatting
				// %v works with any type and shows a reasonable representation
				formattedResponse = fmt.Sprintf("✅ Result: %v", jsonRPCResp.Result)
			} else {
				// Convert JSON to YAML for display
				resultYAML := jsonToYamlForClient(string(resultJSON))
				if resultYAML == "" {
					// If YAML conversion fails, fall back to JSON
					resultJSONPretty, _ := json.MarshalIndent(jsonRPCResp.Result, "", "  ")
					formattedResponse = fmt.Sprintf("✅ Result:\n%s", string(resultJSONPretty))
				} else {
					// Show nicely formatted YAML result
					formattedResponse = fmt.Sprintf("✅ Result:\n%s", resultYAML)
				}
			}
		}

		// Determine success status and error message
		var success bool
		var errorMessage string
		if jsonRPCResp.Error != nil {
			success = false
			errorMessage = jsonRPCResp.Error.Message
		} else {
			success = true
		}

		// Step 9: Return success message to the UI
		// This ResponseMsg will be sent to the Update function in ui.go
		return ResponseMsg{
			Response:     formattedResponse,    // The formatted text to display
			Err:          nil,                  // No error occurred
			RequestData:  string(requestBody),  // Raw request for history
			ResponseData: string(responseBody), // Raw response for history
			StatusCode:   resp.StatusCode,      // HTTP status code
			Success:      success,              // Whether JSON-RPC succeeded
			ErrorMessage: errorMessage,         // Error message if failed
		}
	})
	// The returned function will be executed by bubbletea in a separate goroutine
	// When it completes, bubbletea will send the ResponseMsg to our Update function
}