// Package main contains the user interface logic for the JSON-RPC client
// This file handles all the visual elements, user input, and screen rendering
package main

// Import section - bringing in external libraries we need
import (
	"encoding/json" // JSON encoding and decoding for configuration files
	"fmt"           // Standard library for formatted printing (like printf in C)
	"log"           // For error logging and debugging
	"net/url"       // For URL parsing to create safe directory names
	"os"            // For file operations and environment variables
	"os/exec"       // For executing external commands like Neovim
	"path/filepath" // For working with file paths
	"regexp"        // For regular expressions to sanitize filenames
	"strings"       // For string manipulation and formatting
	"time"          // For timestamps in history files

	// External libraries from GitHub - Go uses module paths as import names
	tea "github.com/charmbracelet/bubbletea" // TUI framework - we alias it as 'tea' for shorter code
	"github.com/charmbracelet/lipgloss"      // CSS-like styling for terminal apps
	"gopkg.in/yaml.v3"                       // YAML parsing and generation library
)

// ServerConfig represents a saved server configuration
type ServerConfig struct {
	// Name is the display name for the server (e.g., "Production API", "Local Dev")
	Name string `json:"name"`
	// URL is the complete server endpoint (e.g., "http://localhost:8080/rpc")
	URL string `json:"url"`
}

// Config represents the complete application configuration
type Config struct {
	// Servers contains the list of predefined server configurations
	Servers []ServerConfig `json:"servers"`
	// LastUsedServer stores the name of the most recently used server
	LastUsedServer string `json:"lastUsedServer,omitempty"`
}

// focusArea is a custom type based on int - this is called "type declaration" in Go
// It represents which part of our UI currently has focus (like which input field is active)
type focusArea int

// Constants in Go are defined with 'const' keyword
// 'iota' is a special Go feature that automatically assigns incremental values (0, 1...)
const (
	focusParams   focusArea = iota // Will be 0 - when parameters input is focused
	focusResponse                  // Will be 1 - when response area is focused
)

// methodModalFocus represents focus areas within the method selection modal
type methodModalFocus int

const (
	focusMethodList   methodModalFocus = iota // Will be 0 - when predefined method list is focused
	focusCustomMethod                         // Will be 1 - when custom method input is focused
)

// serverModalFocus represents focus areas within the server selection modal
type serverModalFocus int

const (
	focusServerList    serverModalFocus = iota // Will be 0 - when server list is focused
	focusCustomServer                          // Will be 1 - when custom server input is focused
	focusServerName                            // Will be 2 - when server name input is focused
)

// EditCompleteMsg is sent when external editor finishes editing
// This custom message carries the edited content back to our Update function
type EditCompleteMsg struct {
	// Content contains the edited text from the external editor
	Content string
	// Err contains any error that occurred during editing (file access, editor not found, etc.)
	Err error
	// Type indicates what was being edited (params, response, headers)
	Type string
}

// ClipboardCompleteMsg is sent when clipboard copy operation finishes
type ClipboardCompleteMsg struct {
	// Success indicates if the copy operation succeeded
	Success bool
	// Err contains any error that occurred during clipboard operation
	Err error
}

// Model is a struct that holds all the data our application needs
// In Go, struct fields that start with capital letters are "exported" (public)
// This means other packages can access them
type Model struct {
	// ServerURL stores the address where we'll send JSON-RPC requests
	ServerURL string

	// Methods is a slice (dynamic array) of strings containing available RPC methods
	// Slices are one of Go's most important data structures
	Methods []string

	// MethodCursor tracks which method is currently selected (like an array index)
	MethodCursor int

	// SelectedMethod stores the currently selected method name
	// This could be from the predefined list or a custom method
	SelectedMethod string

	// ParamsText holds the YAML parameters the user types (displayed as YAML, converted to JSON for requests)
	ParamsText string

	// ResponseText stores the server's response to display (formatted as YAML)
	ResponseText string

	// RawResponseJSON stores the original JSON response for clipboard/editor operations
	RawResponseJSON string

	// Focus tracks which UI component is currently active
	Focus focusArea

	// Loading is a boolean flag - true when we're waiting for server response
	Loading bool

	// RequestID is incremented for each request (JSON-RPC requires unique IDs)
	RequestID int

	// TempServerURL stores the custom server URL being typed in the server modal
	TempServerURL string

	// ShowMethodModal indicates if the method selection modal is currently visible
	// When true, the modal overlays the main interface for method selection
	ShowMethodModal bool

	// MethodModalFocus tracks which component in the method modal has focus
	MethodModalFocus methodModalFocus

	// CustomMethodText stores the custom method name being typed
	// This allows users to enter method names not in the predefined list
	CustomMethodText string

	// RequestHeaders stores HTTP headers for JSON-RPC requests as a JSON string
	// Allows custom headers like authorization, content-type overrides, etc.
	RequestHeaders string

	// Config contains the application configuration loaded from file
	Config Config

	// ShowServerModal indicates if the server selection modal is currently visible
	ShowServerModal bool

	// ServerCursor tracks which server is currently selected in the server modal
	ServerCursor int

	// TempServerName stores the temporary server name when adding a new server
	TempServerName string

	// ServerModalFocus tracks which component in the server modal has focus
	ServerModalFocus serverModalFocus
}

// NewModel is a constructor function that creates and initializes our Model
// In Go, constructor functions are just regular functions that return a struct
// This is the Go way of doing what other languages call "constructors"
func NewModel() Model {
	// Load configuration from file (creates default if doesn't exist)
	config := loadConfig()
	
	// Determine initial server URL from configuration
	var initialServerURL string
	if len(config.Servers) > 0 {
		// Try to use last used server or first in list
		if config.LastUsedServer != "" {
			for _, server := range config.Servers {
				if server.Name == config.LastUsedServer {
					initialServerURL = server.URL
					break
				}
			}
		}
		// Fall back to first server if last used not found
		if initialServerURL == "" {
			initialServerURL = config.Servers[0].URL
		}
	} else {
		// Ultimate fallback if no servers configured
		initialServerURL = "http://localhost:5456/rpc"
	}
	
	// Return a Model struct with initial values
	// This syntax is called "struct literal" - we're creating a struct and setting its fields
	return Model{
		// Setting initial server URL from configuration
		ServerURL: initialServerURL,

		// Creating a slice with predefined method names
		// []string{...} creates a slice of strings
		Methods: []string{
			"ping",        // Simple connectivity test
			"echo",        // Returns what you send it
			"add",         // Mathematical addition
			"subtract",    // Mathematical subtraction
			"multiply",    // Mathematical multiplication
			"divide",      // Mathematical division
			"getStatus",   // Get server status
			"listMethods", // List available methods
		},

		// Start with first method selected (index 0)
		MethodCursor: 0,

		// Set initial selected method to first in list
		SelectedMethod: "ping",

		// Default empty YAML for parameters (will be converted to JSON for requests)
		ParamsText: "",

		// Placeholder text for response area
		ResponseText: "Response will appear here...",

		// Start with parameters field focused (no method selector in main UI)
		Focus: focusParams,

		// Start with request ID 1 (JSON-RPC convention)
		RequestID: 1,

		// TempServerURL starts empty
		TempServerURL: "",

		// Method modal starts hidden
		ShowMethodModal: false,

		// Method modal starts with method list focused
		MethodModalFocus: focusMethodList,

		// Custom method text starts empty
		CustomMethodText: "",

		// Store the loaded configuration
		Config: config,

		// Server modal starts hidden
		ShowServerModal: false,

		// Server modal starts with server list focused
		ServerModalFocus: focusServerList,

		// Server cursor starts at 0
		ServerCursor: 0,

		// Temp server name starts empty
		TempServerName: "",
	}
}

// Init implements the tea.Model interface - this is Go's way of implementing interfaces
// Interfaces in Go are implicit - if a type has the required methods, it implements the interface
// tea.Model interface requires Init(), Update(), and View() methods
func (m Model) Init() tea.Cmd {
	// tea.Cmd represents a command that bubbletea will execute
	// Returning nil means "no initial command to run"
	return nil
}

// Update handles all user input and state changes
// This is the core of the bubbletea architecture - it receives messages and returns new state
// The pattern (Model, tea.Cmd) is common in functional programming
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Type switch - Go's way of checking what type a variable is
	// msg is of type tea.Msg (interface), but could be different concrete types
	switch msg := msg.(type) {

	// tea.KeyMsg represents keyboard input
	case tea.KeyMsg:
		// Check if Method modal is open - handle modal-specific keys
		if m.ShowMethodModal {
			// Method Modal-specific keyboard handling
			switch msg.String() {
			case "escape":
				// Cancel method selection and close modal without saving
				m.ShowMethodModal = false
				m.CustomMethodText = ""

			case "enter":
				// Save the selected method and close modal
				if m.MethodModalFocus == focusMethodList {
					// Use selected predefined method
					m.SelectedMethod = m.Methods[m.MethodCursor]
				} else {
					// Use custom method text
					if strings.TrimSpace(m.CustomMethodText) != "" {
						m.SelectedMethod = strings.TrimSpace(m.CustomMethodText)
					}
				}
				m.ShowMethodModal = false
				m.CustomMethodText = ""

			case "tab":
				// Switch focus between method list and custom input
				if m.MethodModalFocus == focusMethodList {
					m.MethodModalFocus = focusCustomMethod
				} else {
					m.MethodModalFocus = focusMethodList
				}

			case "up", "k":
				// Only move cursor if method list is focused AND we're not at the top
				if m.MethodModalFocus == focusMethodList && m.MethodCursor > 0 {
					m.MethodCursor--
				}

			case "down", "j":
				// Only move cursor if method list is focused AND we're not at the bottom
				if m.MethodModalFocus == focusMethodList && m.MethodCursor < len(m.Methods)-1 {
					m.MethodCursor++
				}

			case "backspace":
				// Delete character from custom method input
				if m.MethodModalFocus == focusCustomMethod && len(m.CustomMethodText) > 0 {
					m.CustomMethodText = m.CustomMethodText[:len(m.CustomMethodText)-1]
				}

			default:
				// Add character to custom method input (single character input)
				if m.MethodModalFocus == focusCustomMethod && len(msg.String()) == 1 {
					m.CustomMethodText += msg.String()
				}
			}
			// Return early - don't process other keys when modal is open
			return m, nil
		}

		// Check if Server modal is open - handle modal-specific keys
		if m.ShowServerModal {
			// Server Modal-specific keyboard handling
			switch msg.String() {
			case "escape":
				// Cancel server selection and close modal without saving
				m.ShowServerModal = false
				m.TempServerName = ""

			case "enter":
				// Select current server or add custom server
				if m.ServerModalFocus == focusServerList && len(m.Config.Servers) > 0 {
					// Select predefined server
					selectedServer := m.Config.Servers[m.ServerCursor]
					m.ServerURL = selectedServer.URL
					m.Config.LastUsedServer = selectedServer.Name
					m.saveConfig() // Save the updated last used server
				} else if m.ServerModalFocus == focusCustomServer && m.TempServerURL != "" {
					// Add and select custom server
					if m.TempServerName == "" {
						// Generate name from URL if no name provided
						m.TempServerName = "Custom Server"
					}
					m.addServerToConfig(m.TempServerName, m.TempServerURL)
					m.ServerURL = m.TempServerURL
					m.Config.LastUsedServer = m.TempServerName
					m.saveConfig()
				}
				m.ShowServerModal = false
				m.TempServerName = ""

			case "tab":
				// Switch focus between server list and custom input
				if m.ServerModalFocus == focusServerList {
					m.ServerModalFocus = focusCustomServer
				} else if m.ServerModalFocus == focusCustomServer {
					m.ServerModalFocus = focusServerName
				} else {
					m.ServerModalFocus = focusServerList
				}

			case "up", "k":
				// Only move cursor if server list is focused AND we're not at the top
				if m.ServerModalFocus == focusServerList && m.ServerCursor > 0 {
					m.ServerCursor--
				}

			case "down", "j":
				// Only move cursor if server list is focused AND we're not at the bottom
				if m.ServerModalFocus == focusServerList && len(m.Config.Servers) > 0 && m.ServerCursor < len(m.Config.Servers)-1 {
					m.ServerCursor++
				}

			case "backspace":
				// Delete character from custom inputs
				if m.ServerModalFocus == focusCustomServer && len(m.TempServerURL) > 0 {
					m.TempServerURL = m.TempServerURL[:len(m.TempServerURL)-1]
				} else if m.ServerModalFocus == focusServerName && len(m.TempServerName) > 0 {
					m.TempServerName = m.TempServerName[:len(m.TempServerName)-1]
				}

			default:
				// Add character to custom inputs (single character input)
				if len(msg.String()) == 1 {
					if m.ServerModalFocus == focusCustomServer {
						m.TempServerURL += msg.String()
					} else if m.ServerModalFocus == focusServerName {
						m.TempServerName += msg.String()
					}
				}
			}
			// Return early - don't process other keys when modal is open
			return m, nil
		}

		// Normal keyboard handling (when no modals are open)
		// Another switch to check which key was pressed
		// msg.String() converts the key press to a string representation
		switch msg.String() {

		// Handle quit commands - user wants to exit
		case "ctrl+c", "q":
			// tea.Quit is a special command that tells bubbletea to exit
			return m, tea.Quit

		// Handle 's' key - show server selection modal
		case "s":
			// Open the server selection modal
			m.ShowServerModal = true
			// Reset server modal focus to server list
			m.ServerModalFocus = focusServerList
			// Clear temporary server name
			m.TempServerName = ""

		// Handle 'm' key - show method selection modal
		case "m":
			// Open the method selection modal
			m.ShowMethodModal = true
			// Start with method list focused
			m.MethodModalFocus = focusMethodList
			// Clear custom method text
			m.CustomMethodText = ""

		// Handle 'e' key - edit parameters in external editor (Neovim/Vim/Nano)
		case "e":
			// Only allow editing if parameters field is focused
			if m.Focus == focusParams {
				// Launch external editor with current parameters content
				// This will suspend our TUI and open the user's preferred editor
				return m, EditInVim(m.ParamsText, "params", false)
			}

		// Handle Tab key - move focus forward through UI components (only 2 now)
		case "tab":
			// Cycle through focus areas using modulo arithmetic
			// (current + 1) % 2 ensures we wrap around from 1 back to 0
			// int(m.Focus) converts our custom type to int for math
			// focusArea(...) converts back to our custom type
			m.Focus = focusArea((int(m.Focus) + 1) % 2)

		// Handle Shift+Tab - move focus backward through UI components
		case "shift+tab":
			// Reverse cycle - adding 2 ensures we don't get negative numbers
			// (current - 1 + 2) % 2 wraps from 0 to 1
			m.Focus = focusArea((int(m.Focus) - 1 + 2) % 2)

		// Handle Enter key - send JSON-RPC request
		case "enter":
			// Send request when Enter is pressed (regardless of focus)
			m.Loading = true
			// MakeRequest is defined in client.go - it returns a tea.Cmd
			// This is Go's way of calling functions from other files in same package
			return m, MakeRequest(m)

		// Handle 'r' key - copy response to system clipboard (as JSON for compatibility)
		case "r":
			if m.RawResponseJSON != "" {
				// Copy raw JSON response to system clipboard for compatibility with other tools
				return m, CopyToClipboard(m.RawResponseJSON)
			} else if m.ResponseText != "" {
				// Fallback to formatted text if no raw JSON available
				return m, CopyToClipboard(m.ResponseText)
			}

		// Handle 'R' key - open server response in vim for viewing (as JSON)
		case "R":
			if m.RawResponseJSON != "" {
				// Open raw JSON response in vim as read-only for better analysis
				return m, EditInVim(m.RawResponseJSON, "response", true)
			} else if m.ResponseText != "" {
				// Fallback to formatted text if no raw JSON available
				return m, EditInVim(m.ResponseText, "response", true)
			}

		// Handle 'k' key - edit request headers in vim
		case "k":
			// Open headers editor (initially empty if no headers set)
			return m, EditInVim(m.RequestHeaders, "headers", false)
		}

	// ResponseMsg is our custom message type (defined in client.go)
	// This handles responses from our JSON-RPC requests
	case ResponseMsg:
		// Request is complete, so stop showing loading indicator
		m.Loading = false

		// Save request/response history to disk (regardless of success/failure)
		// Only save if we have meaningful data to save
		if msg.RequestData != "" || msg.ResponseData != "" {
			m.saveRequestHistory(
				m.SelectedMethod,   // Method name
				msg.RequestData,    // Raw request JSON
				msg.ResponseData,   // Raw response JSON
				msg.StatusCode,     // HTTP status code
				msg.Success,        // Whether JSON-RPC succeeded
				msg.ErrorMessage,   // Error message if failed
			)
		}

		// Check if there was an error with the request
		// In Go, errors are values, not exceptions
		if msg.Err != nil {
			// fmt.Sprintf is like printf - it formats a string with variables
			// %v is a general format specifier that works with any type
			m.ResponseText = fmt.Sprintf("❌ Error: %v", msg.Err)
			m.RawResponseJSON = "" // Clear raw JSON on error
		} else {
			// No error, so display the successful response (already formatted as YAML)
			m.ResponseText = msg.Response
			// Store raw JSON response for clipboard and editor operations
			m.RawResponseJSON = msg.ResponseData
		}

		// Increment request ID for next request (JSON-RPC requires unique IDs)
		m.RequestID++

	// EditCompleteMsg is our custom message type for handling external editor completion
	// This is sent when the user finishes editing in Neovim/Vim and returns to our TUI
	case EditCompleteMsg:
		// Check if there was an error during editing
		if msg.Err != nil {
			// If editing failed, we could show an error, but for now just keep original content
			// The error might be "editor not found" or "file access denied"
			// We'll silently keep the original content so the user can continue
		} else {
			// Editor completed successfully, update the appropriate field based on type
			switch msg.Type {
			case "params":
				// Update parameters with edited content
				m.ParamsText = msg.Content
			case "headers":
				// Update headers with edited content
				m.RequestHeaders = msg.Content
			case "response":
				// Response editing is read-only, so we don't update anything
				// This case exists for completeness but does nothing
			}
		}

	// ClipboardCompleteMsg handles clipboard copy operation results
	case ClipboardCompleteMsg:
		// For now, we silently handle clipboard operations
		// In the future, we could show a status message or notification
		// The operation success/failure is logged but not displayed to avoid UI clutter
	}

	// Return the updated model and no command
	// This is the bubbletea pattern - return new state and any commands to execute
	return m, nil
}

// wrapText wraps text to fit within specified width, preserving words
// This helper function ensures our text fields don't exceed the specified characters per line
// It intelligently breaks on word boundaries to maintain readability
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text) // Split text into words

	for _, word := range words {
		// Check if adding this word would exceed the width
		if currentLine.Len()+len(word)+1 > width {
			// Start a new line
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			// Add word to current line
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Add the last line
	if result.Len() > 0 {
		result.WriteString("\n")
	}
	result.WriteString(currentLine.String())

	return result.String()
}

// EditInVim launches an external editor to edit the given content
// It creates a temporary file, opens it in the user's preferred editor, 
// and returns the edited content when the editor is closed
// Parameters: content (text to edit), editType (params/headers/response), readOnly (if true, opens in read-only mode)
func EditInVim(content, editType string, readOnly bool) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Create a temporary file with .json extension for syntax highlighting
		// os.CreateTemp creates a unique temporary file in the system temp directory
		// Choose file extension based on edit type
		var pattern string
		switch editType {
		case "headers":
			pattern = "govoc-headers-*.json"
		case "response":
			pattern = "govoc-response-*.json"
		default: // params
			pattern = "govoc-params-*.yaml"
		}
		
		tmpfile, err := os.CreateTemp("", pattern)
		if err != nil {
			return EditCompleteMsg{
				Content: content, // Return original content if we can't create temp file
				Err:     fmt.Errorf("failed to create temporary file: %v", err),
				Type:    editType,
			}
		}
		
		// Write the current content to the temporary file
		// This gives the user the existing JSON to edit as a starting point
		_, err = tmpfile.WriteString(content)
		if err != nil {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			return EditCompleteMsg{
				Content: content,
				Err:     fmt.Errorf("failed to write to temporary file: %v", err),
				Type:    editType,
			}
		}
		tmpfile.Close() // Close the file so the editor can open it
		
		// Determine which editor to use
		// Priority: 1) EDITOR environment variable, 2) nvim, 3) vim, 4) nano
		editor := os.Getenv("EDITOR")
		if editor == "" {
			// Try to find nvim first, then fallback to vim, then nano
			if _, err := exec.LookPath("nvim"); err == nil {
				editor = "nvim"
			} else if _, err := exec.LookPath("vim"); err == nil {
				editor = "vim"
			} else if _, err := exec.LookPath("nano"); err == nil {
				editor = "nano"
			} else {
				// No suitable editor found
				os.Remove(tmpfile.Name())
				return EditCompleteMsg{
					Content: content,
					Err:     fmt.Errorf("no suitable editor found (tried: nvim, vim, nano)"),
					Type:    editType,
				}
			}
		}
		
		// Log which editor we're using for debugging
		log.Printf("Using editor: %s to edit file: %s", editor, tmpfile.Name())
		
		// Prepare editor command with appropriate syntax highlighting based on content type
		var cmd *exec.Cmd
		if strings.Contains(editor, "nvim") || strings.Contains(editor, "vim") {
			// Determine filetype based on edit type
			var filetype string
			switch editType {
			case "params":
				filetype = "yaml" // Parameters are edited as YAML
			default:
				filetype = "json" // Headers and response remain JSON
			}
			
			if readOnly {
				// For vim/nvim in read-only mode: set filetype and read-only
				cmd = exec.Command(editor, "-R", fmt.Sprintf("+set ft=%s", filetype), tmpfile.Name())
			} else {
				// For vim/nvim: set appropriate filetype for syntax highlighting
				cmd = exec.Command(editor, fmt.Sprintf("+set ft=%s", filetype), tmpfile.Name())
			}
		} else {
			// For other editors: just open the file (nano doesn't have good read-only mode)
			cmd = exec.Command(editor, tmpfile.Name())
		}
		
		// Connect the editor to the current terminal
		// This allows the editor to take over the screen completely
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// Use tea.ExecProcess to properly manage terminal state
		// This suspends bubbletea, runs the external editor, then resumes bubbletea
		// This prevents terminal conflicts that cause crashes
		c := tea.ExecProcess(cmd, func(err error) tea.Msg {
			// This callback runs after the editor exits
			if err != nil {
				os.Remove(tmpfile.Name())
				return EditCompleteMsg{
					Content: content,
					Err:     fmt.Errorf("editor exited with error: %v", err),
					Type:    editType,
				}
			}
			
			// Read the edited content from the temporary file
			editedBytes, err := os.ReadFile(tmpfile.Name())
			if err != nil {
				os.Remove(tmpfile.Name())
				return EditCompleteMsg{
					Content: content,
					Err:     fmt.Errorf("failed to read edited file: %v", err),
					Type:    editType,
				}
			}
			
			// Clean up the temporary file
			os.Remove(tmpfile.Name())
			
			// Return the edited content to the UI
			return EditCompleteMsg{
				Content: string(editedBytes),
				Err:     nil,
				Type:    editType,
			}
		})
		
		// Return the command that will execute the editor process
		// tea.ExecProcess handles all the terminal state management for us
		return c()
	})
}

// CopyToClipboard copies the given text to the system clipboard
// Uses different commands based on the operating system and available tools
func CopyToClipboard(text string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		var cmd *exec.Cmd
		
		// Try different clipboard commands based on what's available
		if _, err := exec.LookPath("xclip"); err == nil {
			// Linux with xclip
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			// Linux with xsel
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("pbcopy"); err == nil {
			// macOS
			cmd = exec.Command("pbcopy")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			// Wayland (modern Linux)
			cmd = exec.Command("wl-copy")
		} else {
			// No clipboard tool found
			return ClipboardCompleteMsg{
				Success: false,
				Err:     fmt.Errorf("no clipboard tool found (tried: xclip, xsel, pbcopy, wl-copy)"),
			}
		}
		
		// Set up stdin pipe to send text to clipboard command
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return ClipboardCompleteMsg{
				Success: false,
				Err:     fmt.Errorf("failed to create stdin pipe: %v", err),
			}
		}
		
		// Start the clipboard command
		if err := cmd.Start(); err != nil {
			return ClipboardCompleteMsg{
				Success: false,
				Err:     fmt.Errorf("failed to start clipboard command: %v", err),
			}
		}
		
		// Write text to clipboard command's stdin
		_, err = stdin.Write([]byte(text))
		if err != nil {
			stdin.Close()
			return ClipboardCompleteMsg{
				Success: false,
				Err:     fmt.Errorf("failed to write to clipboard: %v", err),
			}
		}
		
		// Close stdin and wait for command to complete
		stdin.Close()
		if err := cmd.Wait(); err != nil {
			return ClipboardCompleteMsg{
				Success: false,
				Err:     fmt.Errorf("clipboard command failed: %v", err),
			}
		}
		
		// Success
		log.Printf("Successfully copied %d characters to clipboard", len(text))
		return ClipboardCompleteMsg{
			Success: true,
			Err:     nil,
		}
	})
}

// getConfigPath returns the path to the configuration file
// Uses XDG_CONFIG_HOME or falls back to ~/.config/govoc/config.json
func getConfigPath() string {
	// Check for XDG_CONFIG_HOME environment variable (Linux standard)
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		// Fall back to ~/.config (standard fallback)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Ultimate fallback to current directory
			return "./govoc-config.json"
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	
	// Create govoc subdirectory and return config file path
	return filepath.Join(configHome, "govoc", "config.json")
}

// loadConfig loads the application configuration from file
// Creates default configuration if file doesn't exist
func loadConfig() Config {
	configPath := getConfigPath()
	
	// Try to read existing configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist or can't be read - create default config
		log.Printf("Could not read config file %s: %v", configPath, err)
		return createDefaultConfig()
	}
	
	// Parse JSON configuration
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Invalid config file %s: %v", configPath, err)
		return createDefaultConfig()
	}
	
	log.Printf("Loaded configuration with %d servers from %s", len(config.Servers), configPath)
	return config
}

// saveConfig saves the current configuration to file
func (m *Model) saveConfig() error {
	configPath := getConfigPath()
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	
	// Marshal configuration to JSON
	data, err := json.MarshalIndent(m.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	// Write configuration file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	
	log.Printf("Saved configuration with %d servers to %s", len(m.Config.Servers), configPath)
	return nil
}

// createDefaultConfig creates a default configuration with example servers
func createDefaultConfig() Config {
	return Config{
		Servers: []ServerConfig{
			{
				Name: "Local Development",
				URL:  "http://localhost:8080/rpc",
			},
			{
				Name: "Local Test Server",
				URL:  "http://localhost:5456/rpc",
			},
		},
		LastUsedServer: "Local Development",
	}
}

// addServerToConfig adds a new server to the configuration
func (m *Model) addServerToConfig(name, url string) {
	// Check if server with this name already exists
	for i, server := range m.Config.Servers {
		if server.Name == name {
			// Update existing server
			m.Config.Servers[i].URL = url
			return
		}
	}
	
	// Add new server
	m.Config.Servers = append(m.Config.Servers, ServerConfig{
		Name: name,
		URL:  url,
	})
}

// HistoryMetadata contains metadata about a JSON-RPC request/response pair
type HistoryMetadata struct {
	// Timestamp when the request was made
	Timestamp time.Time `json:"timestamp"`
	// Server URL that was contacted
	ServerURL string `json:"serverURL"`
	// Method name that was called
	Method string `json:"method"`
	// Request ID from JSON-RPC
	RequestID int `json:"requestID"`
	// HTTP headers that were sent (if any)
	Headers map[string]string `json:"headers,omitempty"`
	// HTTP status code received
	StatusCode int `json:"statusCode"`
	// Whether the JSON-RPC call succeeded (no "error" field in response)
	Success bool `json:"success"`
	// Error message if the call failed
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// getHistoryDir returns the directory path for storing history files
// Uses XDG_DATA_HOME or falls back to ~/.local/share/govoc/history
func getHistoryDir() string {
	// Check for XDG_DATA_HOME environment variable (Linux standard)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		// Fall back to ~/.local/share (standard fallback)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Ultimate fallback to current directory
			return "./govoc-history"
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	
	// Create govoc history subdirectory
	return filepath.Join(dataHome, "govoc", "history")
}

// sanitizeForFilename removes or replaces characters that aren't safe for filenames
func sanitizeForFilename(input string) string {
	// Replace URL schemes and common problematic characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	sanitized := re.ReplaceAllString(input, "-")
	
	// Remove multiple consecutive dashes
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")
	
	// Trim dashes from start and end
	sanitized = strings.Trim(sanitized, "-")
	
	// Ensure we have something if input was all special characters
	if sanitized == "" {
		sanitized = "unknown"
	}
	
	return sanitized
}

// getServerDirName creates a safe directory name from server URL
func getServerDirName(serverURL string) string {
	// Parse URL to get meaningful parts
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		// If URL parsing fails, just sanitize the whole thing
		return sanitizeForFilename(serverURL)
	}
	
	// Use host and port for directory name
	hostPort := parsedURL.Host
	if hostPort == "" {
		hostPort = serverURL // fallback to full URL
	}
	
	return sanitizeForFilename(hostPort)
}

// saveRequestHistory saves a request/response pair to the history directory
func (m *Model) saveRequestHistory(method, requestData, responseData string, statusCode int, success bool, errorMsg string) {
	timestamp := time.Now()
	
	// Create directory structure: history/server/method/history/
	historyBaseDir := getHistoryDir()
	serverDir := getServerDirName(m.ServerURL)
	methodDir := sanitizeForFilename(method)
	
	fullHistoryDir := filepath.Join(historyBaseDir, serverDir, methodDir, "history")
	
	// Create directory structure
	if err := os.MkdirAll(fullHistoryDir, 0755); err != nil {
		log.Printf("Failed to create history directory %s: %v", fullHistoryDir, err)
		return
	}
	
	// Create filename with timestamp
	timeStr := timestamp.Format("2006-01-02T15-04-05")
	baseFilename := fmt.Sprintf("%s-%d", timeStr, m.RequestID)
	
	// Parse headers from RequestHeaders if present
	var headers map[string]string
	if m.RequestHeaders != "" {
		json.Unmarshal([]byte(m.RequestHeaders), &headers)
	}
	
	// Create metadata
	metadata := HistoryMetadata{
		Timestamp:    timestamp,
		ServerURL:    m.ServerURL,
		Method:       method,
		RequestID:    m.RequestID,
		Headers:      headers,
		StatusCode:   statusCode,
		Success:      success,
		ErrorMessage: errorMsg,
	}
	
	// Save metadata file
	metaFile := filepath.Join(fullHistoryDir, baseFilename+"-meta.json")
	if metaData, err := json.MarshalIndent(metadata, "", "  "); err == nil {
		if err := os.WriteFile(metaFile, metaData, 0644); err != nil {
			log.Printf("Failed to save metadata file %s: %v", metaFile, err)
		}
	}
	
	// Save request file
	requestFile := filepath.Join(fullHistoryDir, baseFilename+"-request.json")
	if err := os.WriteFile(requestFile, []byte(requestData), 0644); err != nil {
		log.Printf("Failed to save request file %s: %v", requestFile, err)
	}
	
	// Save response file
	responseFile := filepath.Join(fullHistoryDir, baseFilename+"-response.json")
	if err := os.WriteFile(responseFile, []byte(responseData), 0644); err != nil {
		log.Printf("Failed to save response file %s: %v", responseFile, err)
	}
	
	log.Printf("Saved request history to %s", fullHistoryDir)
}

// jsonToYaml converts JSON string to YAML string for better human readability
func jsonToYaml(jsonStr string) string {
	// Handle empty or whitespace-only input
	if strings.TrimSpace(jsonStr) == "" {
		return ""
	}
	
	// Parse JSON into a generic interface
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		// If JSON parsing fails, return original string
		log.Printf("Failed to parse JSON for YAML conversion: %v", err)
		return jsonStr
	}
	
	// Convert to YAML
	yamlBytes, err := yaml.Marshal(jsonData)
	if err != nil {
		// If YAML marshaling fails, return original string
		log.Printf("Failed to convert to YAML: %v", err)
		return jsonStr
	}
	
	return strings.TrimSpace(string(yamlBytes))
}

// yamlToJson converts YAML string to JSON string for API communication
func yamlToJson(yamlStr string) (string, error) {
	// Handle empty or whitespace-only input
	yamlStr = strings.TrimSpace(yamlStr)
	if yamlStr == "" {
		return "{}", nil // Return empty JSON object for empty YAML
	}
	
	// Parse YAML into a generic interface
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &yamlData); err != nil {
		return "", fmt.Errorf("invalid YAML: %v", err)
	}
	
	// Convert to JSON
	jsonBytes, err := json.Marshal(yamlData)
	if err != nil {
		return "", fmt.Errorf("failed to convert YAML to JSON: %v", err)
	}
	
	return string(jsonBytes), nil
}

// formatJsonAsYaml takes a JSON string and returns it formatted as YAML
// Used for displaying server responses in a more readable format
func formatJsonAsYaml(jsonStr string) string {
	// Handle empty responses
	if strings.TrimSpace(jsonStr) == "" {
		return "No response"
	}
	
	// Try to pretty-format as YAML
	yamlStr := jsonToYaml(jsonStr)
	if yamlStr == "" {
		return jsonStr // Return original if conversion failed
	}
	
	return yamlStr
}

// View renders the user interface to a string
// This function is called every time the screen needs to be redrawn
// It's like the "render" function in React or other UI frameworks
func (m Model) View() string {
	// Define styles using lipgloss - think of it like CSS for the terminal
	// lipgloss.NewStyle() creates a new style object that we can chain methods on

	// Style for the status line at top (shows server URL)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")). // White text color
		Background(lipgloss.Color("#7D56F4")). // Purple background color
		Padding(0, 1).                         // Padding: 0 vertical, 1 horizontal
		Width(120)                             // Fixed width of 120 characters

	// Style for the title bar
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")). // White text color
		Background(lipgloss.Color("#874BFD")). // Darker purple background color
		Padding(0, 1).                         // Padding: 0 vertical, 1 horizontal
		Width(120).                            // Fixed width of 120 characters
		Align(lipgloss.Center)                 // Center the text

	// Style for parameters field when focused (doubled height)
	focusedParamsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).            // Draw a rounded border
		BorderForeground(lipgloss.Color("#874BFD")). // Purple border color
		Padding(1).                                  // 1 unit of padding all around
		Width(120).                                  // Fixed width of 120 characters
		Height(10).                                  // Double height for parameters
		Margin(1, 0)                                 // 1 unit top/bottom margin

	// Style for parameters field when not focused (doubled height)
	unfocusedParamsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).         // Same border shape
		BorderForeground(lipgloss.Color("#666")). // Gray border color
		Padding(1).                               // Same padding
		Width(120).                               // Fixed width of 120 characters
		Height(10).                               // Double height for parameters
		Margin(1, 0)                              // Same margin

	// Style for response field when focused (quadruple height)
	focusedResponseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).            // Draw a rounded border
		BorderForeground(lipgloss.Color("#874BFD")). // Purple border color
		Padding(1).                                  // 1 unit of padding all around
		Width(120).                                  // Fixed width of 120 characters
		Height(20).                                  // Quadruple height for response
		Margin(1, 0)                                 // 1 unit top/bottom margin

	// Style for response field when not focused (quadruple height)
	unfocusedResponseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).         // Same border shape
		BorderForeground(lipgloss.Color("#666")). // Gray border color
		Padding(1).                               // Same padding
		Width(120).                               // Fixed width of 120 characters
		Height(20).                               // Quadruple height for response
		Margin(1, 0)                              // Same margin

	// Style for modal overlays
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).             // Double border for emphasis
		BorderForeground(lipgloss.Color("#FF6B9D")). // Pink border color
		Background(lipgloss.Color("#1A1A1A")).       // Dark background
		Padding(1).                                  // Padding inside modal
		Width(80).                                   // Modal width
		Margin(2, 20)                                // Center the modal

	// Style for method modal components
	methodModalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF88")). // Green border for method modal
		Background(lipgloss.Color("#0A0A0A")).       // Dark background
		Padding(1).
		Width(70).
		Margin(1, 25)

	// Create the status line showing server URL (first line)
	statusLine := statusStyle.Render(fmt.Sprintf("🔗 Server: %s", m.ServerURL))

	// Create the title section
	title := titleStyle.Render("🌐 GOVOC - JSON-RPC Client")

	// Create the method display section (shows current method, click 'm' to change)
	methodDisplayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Width(120).
		Padding(0, 1)

	methodDisplay := methodDisplayStyle.Render(fmt.Sprintf("🎯 Method: %s", m.SelectedMethod))

	// Create the parameters input section with simulated border title
	paramsStyle := unfocusedParamsStyle // Start with unfocused style
	if m.Focus == focusParams {         // If parameters field has focus
		paramsStyle = focusedParamsStyle // Switch to focused style
	}
	
	// Create title label for parameters
	paramsTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")). // White text
		Background(lipgloss.Color("#666")).    // Gray background (matches unfocused border)
		Padding(0, 1).
		Render("📝 Parameters (YAML - use 'e' to edit in vim)")
	
	if m.Focus == focusParams {
		// Change title background to match focused border color
		paramsTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#874BFD")). // Purple background (matches focused border)
			Padding(0, 1).
			Render("📝 Parameters (YAML - use 'e' to edit in vim)")
	}
	
	// Wrap the parameters text to fit within 120 characters
	wrappedParams := wrapText(m.ParamsText, 116) // 116 chars to account for padding/border
	
	// Create the content section
	paramsContent := paramsStyle.Render(wrappedParams)
	
	// Combine title and content vertically with no spacing
	paramsSection := lipgloss.JoinVertical(lipgloss.Left, paramsTitle, paramsContent)

	// Create the response display section with simulated border title
	responseStyle := unfocusedResponseStyle // Start with unfocused style
	if m.Focus == focusResponse {           // If response area has focus
		responseStyle = focusedResponseStyle // Switch to focused style
	}

	// Create title label for response
	responseTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")). // White text
		Background(lipgloss.Color("#666")).    // Gray background (matches unfocused border)
		Padding(0, 1).
		Render("📤 Server Response")
	
	if m.Focus == focusResponse {
		// Change title background to match focused border color
		responseTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#874BFD")). // Purple background (matches focused border)
			Padding(0, 1).
			Render("📤 Server Response")
	}

	// Choose what to display in response area
	responseContent := m.ResponseText // Default: show the actual response
	if m.Loading {                    // If we're waiting for a response
		responseContent = "⏳ Sending request..." // Show loading message instead
	}
	// Wrap the response text to fit within 120 characters
	wrappedResponse := wrapText(responseContent, 116) // 116 chars to account for padding/border
	
	// Create the content section
	responseContentSection := responseStyle.Render(wrappedResponse)
	
	// Combine title and content vertically with no spacing
	responseSection := lipgloss.JoinVertical(lipgloss.Left, responseTitle, responseContentSection)

	// Create help text to show user what keys do what
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")). // Gray text
		Width(120).                         // Fixed width
		Render("💡 Tab/Shift+Tab (navigate) • Enter (send) • e (edit params) • r (copy response) • R (view response) • k (headers) • m (method) • s (server) • q (quit)")

	// Build the main UI components
	mainUI := lipgloss.JoinVertical(
		lipgloss.Left,   // Alignment
		title,           // Title below status
		statusLine,      // Server URL at top
		methodDisplay,   // Current method display
		paramsSection,   // Parameters input
		responseSection, // Response display
		helpText,        // Help text at bottom
	)

	// If Server modal is open, overlay it on top of the main UI
	if m.ShowServerModal {
		// Build server list for modal
		serverListModal := "🌐 Select Server:\n\n"
		
		// Add predefined servers
		for i, server := range m.Config.Servers {
			cursor := " "
			if m.ServerModalFocus == focusServerList && i == m.ServerCursor {
				cursor = ">"
			}
			serverListModal += fmt.Sprintf(" %s %s (%s)\n", cursor, server.Name, server.URL)
		}
		
		serverListModal += "\n"
		
		// Add custom server input section
		serverCursor := " "
		if m.ServerModalFocus == focusCustomServer {
			serverCursor = ">"
		}
		serverListModal += fmt.Sprintf(" %s Custom URL: %s\n", serverCursor, m.TempServerURL)
		
		nameCursor := " "
		if m.ServerModalFocus == focusServerName {
			nameCursor = ">"
		}
		serverListModal += fmt.Sprintf(" %s Name: %s\n", nameCursor, m.TempServerName)
		
		serverListModal += "\nTab: Navigate • Enter: Select • Escape: Cancel"
		
		modal := modalStyle.Render(serverListModal)

		// Overlay modal on main UI
		return lipgloss.Place(120, 25, lipgloss.Center, lipgloss.Center, mainUI) + "\n" +
			lipgloss.Place(120, 25, lipgloss.Center, lipgloss.Center, modal)
	}

	// If Method modal is open, overlay it on top of the main UI
	if m.ShowMethodModal {
		// Build method list for modal
		methodListModal := "📋 Select Method:\n\n"
		for i, method := range m.Methods {
			cursor := "  " // Default: no cursor
			if i == m.MethodCursor && m.MethodModalFocus == focusMethodList {
				cursor = "▶ " // Show arrow cursor
			}
			methodListModal += fmt.Sprintf("%s%s\n", cursor, method)
		}

		// Custom method input section
		customFocus := ""
		if m.MethodModalFocus == focusCustomMethod {
			customFocus = "▶ "
		}
		customMethodSection := fmt.Sprintf("\n%sCustom: %s", customFocus, m.CustomMethodText)

		// Complete modal content
		modalContent := methodListModal + customMethodSection +
			"\n\nTab: Switch • ↑/↓: Navigate • Enter: Select • Escape: Cancel"

		modal := methodModalStyle.Render(modalContent)

		// Overlay modal on main UI
		return lipgloss.Place(120, 25, lipgloss.Center, lipgloss.Center, mainUI) + "\n" +
			lipgloss.Place(120, 25, lipgloss.Center, lipgloss.Center, modal)
	}

	// Return the main UI without any modals
	return mainUI
}
