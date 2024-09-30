# Edigo

Edigo is a collaborative text editor that allows real-time editing across multiple users. It utilizes Conflict-free Replicated Data Types (CRDTs) to ensure consistency across all instances.

## Features

- Real-time collaborative editing
- Network-based session sharing
- Syntax highlighting for various programming languages
- Built-in file management

## Technology Stack

- Go (Golang) for backend logic
- CRDT for conflict resolution
- UDP for session discovery
- TCP for data synchronization

## Main Components

1. **Editor**: Manages the text editing functionality and user interface.
2. **CRDT**: Handles data consistency and conflict resolution.
3. **Network**: Manages session creation, discovery, and data transmission.
4. **UI**: Provides the user interface for the editor.

## How to Use

1. Run the application: `go run ./cmd/main.go [filename]`
2. Use the menu to create or join a collaborative session.
3. Edit the file collaboratively in real-time.

## Development

To contribute to the project:

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request
