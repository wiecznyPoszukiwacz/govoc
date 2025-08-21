// Package main is the entry point for our JSON-RPC client application
// This file orchestrates the startup of our TUI application
// It's the smallest file because it just ties everything together
package main

// Import section - we need these libraries to start our application
import (
	"fmt" // For printing error messages to the console
	"os"  // For exiting the program with error codes

	// Import bubbletea framework for creating terminal user interfaces
	// We alias it as 'tea' to make our code shorter and more readable
	tea "github.com/charmbracelet/bubbletea"
)

// main is the entry point function for any Go program
// When you run "go run .", this is the first function that gets called
// Every Go program must have exactly one main() function in package main
func main() {
	// Step 1: Create and initialize our application model
	// NewModel() is defined in ui.go - it returns a Model struct
	// with all the initial state (server URL, methods list, etc.)
	// This is like creating the "starting state" of our application
	initialModel := NewModel()

	// Step 2: Create a bubbletea Program with our model
	// tea.NewProgram() creates a new TUI program instance
	// It takes our model and configuration options
	//
	// tea.WithAltScreen() is an important option:
	// - Without it: our app writes to the terminal like a normal command
	// - With it: our app takes over the entire terminal screen
	//   When the app exits, it restores the previous terminal content
	//   This gives us a full-screen app experience like vim or htop
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	// Step 3: Run the program and handle any startup errors
	// p.Run() starts the event loop and blocks until the program exits
	// It returns two values:
	// - final model state (we ignore this with _)
	// - error (if something went wrong during startup or execution)
	//
	// The pattern "if var, err := func(); err != nil" is very common in Go
	// It combines assignment and error checking in one line
	if _, err := p.Run(); err != nil {
		// If there was an error, print it to standard error output
		// fmt.Printf works like printf in C - %v means "format this value appropriately"
		fmt.Printf("Error running application: %v", err)

		// Exit with code 1 to indicate an error occurred
		// Exit code 0 means success, non-zero means error
		// This is important for scripts and process managers
		os.Exit(1)
	}

	// If we reach here, the program exited successfully
	// Go functions can end without explicit return statement
	// The program will exit with code 0 (success) automatically
}

// Note for beginners:
// This main.go file demonstrates several important Go concepts:
//
// 1. Package structure: main package + main() function = executable program
// 2. Import organization: standard library first, then external packages
// 3. Error handling: the "if err != nil" pattern is everywhere in Go
// 4. Function calls: calling functions from other files in the same package
// 5. Configuration: using options like tea.WithAltScreen() to modify behavior
//
// The separation of concerns is clear:
// - main.go: program startup and error handling
// - ui.go: user interface logic and state management  
// - client.go: network communication and JSON-RPC protocol
//
// This makes the code easier to understand, test, and maintain!