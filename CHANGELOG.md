# Changelog

All notable changes to this project will be documented in this file.

## [4.1.0] - 2025-08-21

### Major UX Improvements
- **YAML-First Interface**: Parameters are now displayed and edited in human-friendly YAML format
- **YAML Response Display**: Server responses shown in YAML for improved readability
- **Smart Format Conversion**: YAML automatically converted to JSON for API communication
- **Dual Format Support**: UI shows YAML while maintaining JSON compatibility for tools

### Enhanced Editor Integration
- **YAML Syntax Highlighting**: Parameter editing in vim now uses YAML syntax highlighting
- **Format-Specific File Extensions**: Temporary files use `.yaml` for params, `.json` for headers/response
- **Intelligent Editor Commands**: Vim automatically sets appropriate filetype (yaml/json)

### Improved Compatibility 
- **JSON Clipboard Copy**: `r` key copies raw JSON response to clipboard for tool compatibility
- **JSON Editor View**: `R` key opens response in vim as JSON for detailed analysis
- **Preserved History Format**: Request/response history remains in JSON for API compatibility

### Technical Implementation
- **gopkg.in/yaml.v3**: Added YAML parsing and generation library
- **Bidirectional Conversion**: Seamless YAML↔JSON conversion with error handling
- **Dual Response Storage**: UI displays YAML while preserving raw JSON for operations
- **Fallback Mechanisms**: Graceful handling of conversion failures

### User Experience Benefits
```yaml
# Before (JSON)
{"user": {"name": "John", "settings": {"theme": "dark", "notifications": true}}}

# After (YAML) - Much more readable!
user:
  name: John
  settings:
    theme: dark
    notifications: true
```

## [4.0.0] - 2025-08-21

### Major Features Added
- **Server Configuration Management**: Load server list from `~/.config/govoc/config.json` with name-URL pairs
- **Server Selection Modal**: Press 's' to choose from predefined servers or add custom servers to config
- **Automatic Request History**: Every request/response automatically saved to `~/.local/share/govoc/history/`
- **Hierarchical History Structure**: Organized by server → method → history with timestamp-based files
- **Complete Request Tracking**: Each request saves metadata (timestamp, status, success), raw request JSON, and raw response JSON

### Changed
- **Breaking**: Server selection moved from simple URL modal to comprehensive server management
- **Enhanced ResponseMsg**: Extended with history tracking data (request/response/status/success)
- **Configuration-driven**: Initial server URL loaded from configuration file with persistence

### Technical Implementation
- **XDG Base Directory Support**: Follows Linux standards for config (~/.config) and data (~/.local/share)
- **Safe Filename Sanitization**: URL-to-directory conversion with regex-based character filtering
- **Comprehensive History Metadata**: Includes timestamps, HTTP status codes, JSON-RPC success status, error messages
- **Automatic Directory Creation**: History directories created on-demand with proper permissions
- **Cross-platform Paths**: Uses filepath.Join for OS-appropriate path handling

### File Structure Created
```
~/.config/govoc/config.json          # Server configurations
~/.local/share/govoc/history/         # Request history
├── localhost-8080/                   # Server directory
│   ├── ping/                         # Method directory
│   │   └── history/                  # History files
│   │       ├── 2025-08-21T10-30-15-1-meta.json     # Metadata
│   │       ├── 2025-08-21T10-30-15-1-request.json  # Request
│   │       └── 2025-08-21T10-30-15-1-response.json # Response
│   └── getUserInfo/...
└── production-api/...
```

## [3.3.0] - 2025-08-21

### Added
- **Clipboard integration**: Press 'r' to copy server response to system clipboard (supports xclip, xsel, pbcopy, wl-copy)
- **Response viewer**: Press 'R' to open server response in vim as read-only for better viewing
- **Headers editor**: Press 'k' to edit custom HTTP headers in vim
- **Enhanced field sizes**: Parameters field doubled in height, response field quadrupled for better content visibility

### Changed
- **Disabled in-app editing**: Removed direct text input in parameters field - now vim-only editing via 'e' key
- **Updated help text**: Reflects new keyboard shortcuts and vim-based workflow
- **Improved editor integration**: Enhanced EditInVim function supports different content types and read-only mode

### Technical Details
- Added ClipboardCompleteMsg for clipboard operation feedback
- Extended EditCompleteMsg with Type field to handle params/headers/response
- Created separate styles for parameters and response fields with fixed heights
- Cross-platform clipboard support with fallback detection

## [3.2.1] - 2025-08-21

### Fixed
- **External editor crash**: Fixed program crash when opening Neovim/Vim for parameter editing
- **Terminal state management**: Replaced direct `cmd.Run()` with `tea.ExecProcess()` for proper terminal handling
- **Silent crashes**: Added debug logging to help identify editor-related issues

### Technical Details
- Used `tea.ExecProcess()` to properly suspend bubbletea before launching external editor
- Added comprehensive error handling and logging for editor operations
- Fixed terminal state conflicts between bubbletea's alternate screen mode and external editors
- Improved callback-based editor completion handling

## [3.2.0] - 2025-08-21

### Added
- **External editor integration**: Press 'e' while in parameters field to edit JSON in Neovim/Vim/Nano
- **Smart editor detection**: Automatically detects available editors (EDITOR env var → nvim → vim → nano)
- **JSON syntax highlighting**: Sets proper filetype for vim/nvim to enable syntax highlighting
- **Temporary file handling**: Creates secure temporary files for editing with automatic cleanup
- **Editor preference support**: Respects $EDITOR environment variable for custom editor choice

### Enhanced
- **Professional JSON editing**: Full editor capabilities including plugins, autocomplete, and advanced features
- **Seamless workflow**: Editor launches, suspends TUI, returns edited content automatically
- **Error handling**: Graceful fallback if editor not found or editing fails
- **Help text updated**: Added 'e (edit)' to keyboard shortcuts display

### Technical Details
- Implemented `EditInNeovim()` function with comprehensive error handling
- Added `EditCompleteMsg` custom message type for async editor communication
- Smart editor command construction with JSON filetype hints
- Proper terminal state management during external editor execution
- Secure temporary file creation with `.json` extension for proper syntax highlighting

## [3.1.0] - 2025-08-21

### Improved
- **Border titles**: Parameters and response field labels now appear on the border frame
- **Space efficiency**: Gained 2 additional content lines in each field by moving labels to borders
- **Cleaner layout**: Removed internal field labels, creating more professional appearance
- **Better visual hierarchy**: Title and server URL positions swapped for improved information flow
- **Method display**: Removed border around method display for cleaner presentation

### Technical Details
- Implemented simulated border titles using lipgloss composition for backward compatibility
- Used `JoinVertical()` to combine title bars with content sections
- Dynamic title styling that changes color based on focus state (gray/purple background)
- Maintained focus state styling while enhancing label presentation

## [3.0.0] - 2025-08-21

### Major Changes
- **BREAKING**: Method selection moved to modal dialog activated by 'm' key
- Increased UI width from 80 to 120 characters for better content display
- Parameters field now supports multiline input with Ctrl+J for newlines

### Added
- Modal dialog for method selection with predefined methods list
- Custom method input field in method modal for entering any RPC method name
- Multiline JSON parameter editing with intelligent text wrapping
- Enhanced keyboard controls for method modal (Tab to switch focus, ↑/↓ to navigate)
- 120-character wide layout for better readability and content capacity
- Method display in main UI showing current selected method

### Changed
- Main UI layout redesigned without method selector (now modal-only)
- Parameters field supports complex multiline JSON structures
- Focus management simplified to parameters and response fields only
- Text wrapping adjusted for 120-character width (116 chars content + borders)
- Help text updated to reflect new keyboard shortcuts

### Improved
- Better workflow separation between method selection and parameter editing
- More screen space for complex JSON parameters and responses
- Enhanced modal system with proper focus isolation
- Cleaner main interface with dedicated method configuration workflow

### Technical Details
- Added `methodModalFocus` type for managing focus within method modal
- Implemented `SelectedMethod` field to store current method independently
- Enhanced keyboard handling for multiline text input
- Modal overlay system supports multiple concurrent modals

## [2.3.0] - 2025-08-21

### Added
- Modal dialog for server URL editing triggered by 's' key
- 80-character width constraint for parameters and response fields
- Text wrapping functionality to maintain readability within width limits
- Server URL display in top status line for constant visibility
- Enhanced keyboard controls for modal (Enter: save, Escape: cancel)

### Changed
- Redesigned UI layout with server URL moved to first line
- Removed URL field from normal focus cycle (now modal-only)
- Parameters and response fields now wrap text at 80 characters
- Improved visual hierarchy with status line, title, and content sections
- Enhanced help text to include server editing shortcut

### Improved
- Better space utilization with fixed-width layout
- Modal overlay system with proper focus management
- More intuitive server URL editing workflow
- Consistent 80-character terminal layout following conventions

## [2.2.0] - 2025-08-21

### Added
- Comprehensive beginner-friendly documentation throughout all source files
- Detailed line-by-line comments explaining Go concepts and patterns
- Educational explanations of bubbletea architecture and message passing
- In-depth JSON-RPC protocol implementation documentation
- Step-by-step breakdown of network request handling
- Extensive comments on Go-specific features (interfaces, slices, error handling)
- Code examples and explanations for beginners learning Go programming

### Documentation Highlights
- **ui.go**: Complete explanation of TUI architecture, event handling, and styling
- **client.go**: Detailed walkthrough of HTTP requests, JSON parsing, and async operations  
- **main.go**: Foundation concepts of Go program structure and error handling
- All comments written for developers new to Go programming language

## [2.1.0] - 2025-08-21

### Changed
- Refactored project structure with modular architecture
- Moved all source code to `src/` directory
- Separated UI logic into `src/ui.go` module
- Separated JSON-RPC client logic into `src/client.go` module
- Created clean `src/main.go` entry point
- Updated package.json scripts to work with new structure
- Improved code organization and maintainability

### Technical Details
- UI module handles all interface rendering, styling, and user input
- Client module manages JSON-RPC protocol, HTTP requests, and response formatting  
- Main module orchestrates application initialization and program execution
- Each module has clear separation of concerns for better testing and extension

## [2.0.0] - 2025-08-21

### Added
- Complete JSON-RPC 2.0 client functionality
- HTTP POST request handling with proper error management
- Method selector with predefined RPC methods (ping, echo, add, subtract, multiply, divide, getStatus, listMethods)
- Large text area for JSON parameters input with real-time editing
- Server response display with formatted JSON output
- Tab/Shift+Tab navigation between UI components
- Visual focus indicators using lipgloss styling
- Async request handling with loading states
- Comprehensive error handling for network, JSON parsing, and RPC errors
- Server URL configuration (default: http://localhost:8080/rpc)

### Changed
- Completely redesigned TUI interface for JSON-RPC workflow
- Enhanced keyboard controls for multi-component navigation
- Improved styling with lipgloss for better visual hierarchy

## [1.0.0] - 2025-08-21

### Added
- Initial project setup with Go modules
- Basic TUI interface using bubbletea library
- Interactive menu with navigation and selection
- NerdFonts icons in menu items
- Keyboard navigation (↑/↓, j/k)
- Selection functionality (space/enter)
- Quit functionality (q, ctrl+c)
- Package.json for project metadata and scripts