# Parsec: Terminal-Based File Summarizer TUI

Parsec is a fast, terminal-based file inspector that provides immediate summaries of source code files. Built for developers who need rapid context-aware code browsing.

## Features

- **Split-screen interface**: File tree on the left, detailed summary on the right
- **Directory navigation**: Browse through subdirectories like a file manager
- **Directory preview**: Shows contents of selected directories in the summary pane
- **Fuzzy search**: Quickly find files with real-time fuzzy matching (press `/`)
- **Multi-language support**: Go, Python, JavaScript, TypeScript, Rust, Java, C/C++, and more
- **Enhanced file parsing**: 
  - **Markdown rendering**: Beautiful terminal markdown with syntax highlighting
  - **Configuration files**: JSON, YAML, INI, ENV, and more
  - **Text files**: Preview with line counts and content display
  - **Executable help**: Automatically shows help output for executables
- **Intelligent parsing**: Extracts functions, imports, types, and structs
- **Rich file icons**: Visual file type indicators for over 50 file types
- **Responsive design**: Automatically adapts to terminal resize events
- **Fast navigation**: Keyboard-driven interface with vim-style keybindings
- **Async operations**: Non-blocking file summarization for smooth UX

## Installation

```bash
git clone https://github.com/Cod-e-Codes/parsec.git
cd parsec
go build .
```

## Usage

```bash
# Scan current directory
./parsec

# Scan specific directory
./parsec /path/to/project

# Windows examples
./parsec "C:\Users\username\Projects"
./parsec "C:\Program Files\MyApp"

# Unix examples  
./parsec /home/user/code
./parsec ~/Documents/projects

# Show help
./parsec -h
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `k/j` | Navigate file list |
| `Enter` | Enter directory or open file |
| `/` | Start fuzzy search |
| `PgUp/PgDn` | Scroll summary content |
| `Home/End` | Jump to first/last file |
| `t` | Toggle directory visibility |
| `r` | Refresh current directory |
| `q` or `Ctrl+C` | Quit |

### Fuzzy Search Mode

| Key | Action |
|-----|--------|
| `Type` | Add characters to search query |
| `Backspace` | Remove last character |
| `Enter` | Confirm search and stay filtered |
| `ESC` | Cancel search and show all files |

## Fuzzy Search

Parsec includes powerful fuzzy search capabilities:

- **Real-time filtering**: Files are filtered as you type, with instant results
- **Fuzzy matching**: Find files even with typos or partial names (e.g., `mago` matches `main.go`)
- **Smart scoring**: Files are ranked by relevance, with better matches appearing first
- **Visual feedback**: Search query and cursor are displayed at the bottom
- **Easy control**: Enter search mode with `/`, exit with `ESC`, or confirm with `Enter`

### Example Usage
- Press `/` to start searching
- Type `mai` to find `main.go`, `Makefile`, etc.
- Type `sum.go` to find `ui/summary.go`
- Use `Backspace` to refine your search
- Press `Enter` to keep the filtered results
- Press `ESC` to show all files again

## Directory Navigation

Parsec supports intuitive directory navigation with live preview:

- **Current path display**: Shows your current location at the top
- **Directory preview**: Hover over any directory to see its contents in the summary pane
- **Enter directories**: Press `Enter` on a directory to navigate into it
- **Go up**: Use the `..` entry (marked with ⬆️) to go up one level
- **Breadcrumb navigation**: Current directory path is displayed relative to starting point
- **Smart filtering**: Only shows source files and relevant directories
- **Content statistics**: See file/directory counts and listings before entering

## Architecture

### Core Components

- **File Walker**: Lists directory contents and identifies file types
- **Summarizer Engine**: Parses files using language-specific regex patterns  
- **TUI Components**: Split-pane interface with responsive layout
- **Async Worker**: Background file processing for smooth interaction

### Supported File Types

| Category | Extensions | Features Detected |
|----------|------------|-------------------|
| **Programming Languages** | | |
| Go | `.go` | Functions, types, structs, imports |
| Python | `.py` | Functions, classes, imports |
| JavaScript | `.js` | Functions, classes, imports |
| TypeScript | `.ts` | Functions, interfaces, types, imports |
| Rust | `.rs` | Functions, structs, enums, uses |
| Java | `.java` | Functions, classes, imports |
| C/C++ | `.c`, `.cpp`, `.h`, `.hpp` | Functions, structs, includes |
| **Documentation** | | |
| Markdown | `.md`, `.markdown` | Headers, links, **rendered content** |
| Text | `.txt`, `.rst` | Content preview, line count |
| **Configuration** | | |
| JSON | `.json` | Configuration keys, structure |
| YAML | `.yaml`, `.yml` | Configuration keys, structure |
| INI/Config | `.ini`, `.cfg`, `.conf` | Sections, configuration keys |
| Environment | `.env` | Environment variables |
| **Data Files** | | |
| XML | `.xml` | Content preview, structure |
| CSV | `.csv` | Content preview, data format |
| Log Files | `.log` | Content preview, line count |
| **Executables** | `.exe`, `.bin`, etc. | **Automatic help display** |

## Project Structure

```
parsec/
├── main.go             # Application entrypoint and UI coordination
├── ui/
│   ├── filelist.go     # File navigation component
│   └── summary.go      # Summary display component
├── core/
│   └── summarize.go    # File parsing and analysis logic
├── utils/
│   └── filewalker.go   # Filesystem traversal utilities
└── README.md
```

## Recent Improvements

### Responsive Layout (v1.1)
- Fixed terminal resize handling inconsistencies
- Improved component dimension calculations
- Added minimum size constraints for small terminals
- Enhanced scroll position management during resize events

### Stable Scrolling (v1.2)
- Fixed scrolling behavior to only move content within borders
- Eliminated border/frame shifting during scroll operations
- Implemented fixed-height content areas with padding
- Improved visual stability during navigation

### Directory Navigation (v1.3)
- Added support for browsing through subdirectories
- Implemented breadcrumb navigation with current path display  
- Added parent directory navigation with ".." entries
- Enhanced file list to show only current directory contents
- Improved Enter key functionality for directory traversal

### Enhanced File Parsing (v1.4)
- **Markdown rendering**: Beautiful terminal markdown using Glamour with syntax highlighting
- **Configuration file support**: JSON, YAML, INI, ENV with key extraction
- **Text file previews**: Content display with line counts for all text-based files
- **Executable help**: Automatic help command detection and display for executables
- **Rich file icons**: Over 50 file type icons including programming languages, configs, and data files
- **Improved content display**: Better formatting for different file types with appropriate sections

### Fuzzy Search (v1.5)
- **Real-time search**: Instant file filtering as you type with `/` key
- **Fuzzy matching**: Smart algorithm finds files even with partial or mistyped names
- **Search persistence**: Option to keep filtered results or return to full list
- **Visual feedback**: Clear search interface with query display and cursor
- **Enhanced navigation**: Quickly jump to any file in large directories

### Performance Optimizations
- Async file summarization prevents UI blocking
- Efficient regex-based parsing for multiple languages
- Smart directory filtering to skip non-source directories
- Optimized fuzzy search with intelligent ranking

### Code Quality Improvements (v1.6)
- **Codebase cleanup**: Removed unused methods and duplicate code for better maintainability
- **Consolidated parsers**: Simplified file type handling by consolidating similar parsers
- **Extracted common logic**: Unified file selection and summarization logic
- **Improved constants**: Centralized UI dimension constants to eliminate duplication

### Directory Preview (v1.7)
- **Smart directory inspection**: When selecting a directory, shows its contents in the summary pane
- **File and folder counts**: Displays statistics about directory contents (files vs directories)
- **Rich file listings**: Shows up to 20 items with appropriate file type icons
- **Empty directory handling**: Special message for empty directories with navigation hints
- **Parent directory info**: Clear indication when selecting ".." with destination path
- **Instant feedback**: Async loading prevents UI blocking for large directories

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with various file types and terminal sizes
5. Submit a pull request

## License

MIT License - see LICENSE file for details. 