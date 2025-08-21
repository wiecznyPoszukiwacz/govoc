# GOVOC - JSON-RPC TUI Client

[![Version](https://img.shields.io/badge/version-4.1.0-blue.svg)](https://github.com/govoc/govoc)
[![Go](https://img.shields.io/badge/go-1.23+-green.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A modern, feature-rich JSON-RPC client with a beautiful Terminal User Interface (TUI) built with [bubbletea](https://github.com/charmbracelet/bubbletea). GOVOC provides a YAML-first experience for better readability while maintaining full JSON-RPC compatibility.

## ✨ Features

### 🎨 YAML-First Interface
- **Human-friendly parameter input** - Write parameters in clean YAML format
- **Readable response display** - Server responses formatted as YAML
- **Smart conversion** - Automatic YAML ↔ JSON conversion for API communication
- **Syntax highlighting** - Full vim integration with appropriate highlighting

### 🔧 Professional Editor Integration
- **Vim/Neovim support** - Edit parameters, headers, and view responses
- **Read-only response viewer** - Analyze responses in your favorite editor
- **Custom headers editor** - Add authentication, content-type, etc.
- **Temporary file management** - Secure temporary files with proper cleanup

### 🌐 Server Management
- **Configuration-based** - Manage servers in `~/.config/govoc/config.json`
- **Quick server switching** - Press `s` to select from predefined servers
- **Custom server addition** - Add new servers on the fly
- **Last-used persistence** - Remembers your preferred server

### 📂 Complete Request History
- **Automatic logging** - Every request/response saved automatically
- **Hierarchical organization** - `server/method/history/` structure
- **Rich metadata** - Timestamps, HTTP status, success/failure tracking
- **XDG compliance** - Follows Linux standards for config and data directories

### 📋 Cross-Platform Clipboard
- **Smart copy** - Raw JSON copied for tool compatibility
- **Multiple clipboard backends** - xclip, xsel, pbcopy, wl-copy support
- **One-key operation** - Press `r` to copy response

### ⌨️ Intuitive Keyboard Shortcuts
- **Navigation** - Tab/Shift+Tab to move between fields
- **Actions** - Single-key shortcuts for all operations
- **Modal interfaces** - Clean server and method selection
- **Vim-style keys** - Familiar j/k navigation where applicable

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/your-repo/govoc
cd govoc

# Build the application
npm run build

# Or run directly
npm start
```

### First Run

1. **Launch GOVOC** - `./govoc` or `npm start`
2. **Select server** - Press `s` to choose or add a server
3. **Choose method** - Press `m` to select JSON-RPC method
4. **Edit parameters** - Press `e` to write parameters in YAML
5. **Send request** - Press `Enter` to execute
6. **View response** - Response appears as readable YAML

## 📖 Usage Guide

### Parameter Editing (YAML Format)

Instead of complex JSON, write clean YAML:

```yaml
# Simple parameters
user_id: 123
include_profile: true

# Complex nested structures
user:
  name: John Doe
  settings:
    theme: dark
    notifications:
      email: true
      push: false
    preferences:
      - feature_a
      - feature_b
```

### Server Configuration

GOVOC automatically creates `~/.config/govoc/config.json`:

```json
{
  "servers": [
    {
      "name": "Local Development",
      "url": "http://localhost:8080/rpc"
    },
    {
      "name": "Production API",
      "url": "https://api.example.com/rpc"
    }
  ],
  "lastUsedServer": "Local Development"
}
```

### Request History

Every request is automatically saved to `~/.local/share/govoc/history/`:

```
~/.local/share/govoc/history/
├── localhost-8080/
│   ├── getUserInfo/
│   │   └── history/
│   │       ├── 2025-08-21T14-30-15-1-meta.json     # Request metadata
│   │       ├── 2025-08-21T14-30-15-1-request.json  # Raw request
│   │       └── 2025-08-21T14-30-15-1-response.json # Raw response
│   └── createUser/
└── api-example-com/
```

## ⌨️ Keyboard Shortcuts

| Key | Action | Description |
|-----|--------|-------------|
| `Tab` / `Shift+Tab` | Navigate | Move between UI sections |
| `Enter` | Send Request | Execute JSON-RPC call |
| `e` | Edit Parameters | Open parameters in vim (YAML format) |
| `r` | Copy Response | Copy raw JSON response to clipboard |
| `R` | View Response | Open response in vim (JSON format) |
| `k` | Edit Headers | Edit custom HTTP headers in vim |
| `m` | Select Method | Choose JSON-RPC method from list |
| `s` | Select Server | Choose server or add new one |
| `q` / `Ctrl+C` | Quit | Exit application |

## 🏗️ Architecture

### YAML ↔ JSON Flow

```
User Input (YAML) → Internal Storage → JSON-RPC → Server
                                    ↓
UI Display (YAML) ← Format Convert ← JSON Response
                                    ↓
History/Clipboard (JSON) ← Raw Preserve
```

### Directory Structure

```
govoc/
├── src/
│   ├── main.go      # Application entry point
│   ├── ui.go        # TUI interface logic
│   └── client.go    # JSON-RPC client implementation
├── package.json     # Project metadata and scripts
├── go.mod          # Go module dependencies
├── CHANGELOG.md    # Version history
└── README.md       # This file
```

## 🔧 Configuration

### Environment Variables

- `EDITOR` - Preferred text editor (default: nvim → vim → nano)
- `XDG_CONFIG_HOME` - Configuration directory (default: `~/.config`)
- `XDG_DATA_HOME` - Data directory (default: `~/.local/share`)

### Config File Location

- Linux: `~/.config/govoc/config.json`
- macOS: `~/.config/govoc/config.json`
- Windows: `%APPDATA%/govoc/config.json`

## 🛠️ Development

### Requirements

- Go 1.23 or higher
- npm (for build scripts)
- Text editor (vim/nvim recommended)

### Building

```bash
# Install dependencies
go mod tidy

# Build application
npm run build

# Run tests
npm test

# Format code
npm run fmt

# Lint code
npm run vet
```

## 📚 Examples

### Simple Method Call

```yaml
# Parameters (YAML input)
id: 42
```

```json
// Generated JSON-RPC request
{
  "jsonrpc": "2.0",
  "method": "getUserInfo",
  "params": {"id": 42},
  "id": 1
}
```

### Complex Nested Parameters

```yaml
# Parameters (YAML input)
user:
  name: Alice Johnson
  email: alice@example.com
  settings:
    notifications:
      email: true
      sms: false
    theme: dark
    language: en
filters:
  - active
  - verified
include:
  profile: true
  activity: false
```

## 🐛 Troubleshooting

### Common Issues

**Editor not opening:**
- Ensure vim/nvim is installed: `which nvim`
- Set EDITOR environment variable: `export EDITOR=vim`

**Clipboard not working:**
- Install clipboard tool: `sudo apt install xclip` (Linux)
- Verify tool works: `echo "test" | xclip -selection clipboard`

**Config file not found:**
- First run creates default config automatically
- Check permissions: `ls -la ~/.config/govoc/`

**Invalid YAML errors:**
- Check indentation (use spaces, not tabs)
- Validate YAML syntax online before pasting
- Use quotes for strings with special characters

### Debug Mode

Run with verbose logging:
```bash
./govoc 2> govoc.log  # Capture logs to file
tail -f govoc.log     # Monitor in another terminal
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [bubbletea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing and generation
- JSON-RPC 2.0 specification

## 🔗 Links

- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [YAML Specification](https://yaml.org/spec/)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

---

**Made with ❤️ using Go and bubbletea**
