# AGENTS.md - Swings Development Agent

You are responsible for developing the **swings** project—a Go backend service providing MCP (Model Context Protocol) and HTTP interfaces.

## Context & Architecture
- **Architecture Overview**: Refer to `@README.md` and `@docs/SYS_ARCH.md` for project structure and system architecture.
- **MCP Protocol Spec**: Follow version `2026-07-28` with stateless sessions.
- **Backend Module**: `src/backend`
- **Main Entry Point**: `src/backend/cmd/swings/swings.go`

## Scope & File Operations

### Directory Inspection
- **General Layout**: Rely on `@README.md` for the directory layout.
- **Directory Search**: Limit directory listing tools (e.g., `list_directory`) strictly to `src/backend/`.
- **Excluded Paths**: Ignore `.task/`, `axe/`, and `dist/`.

### File Editing & Reading
- **Documentation**: Edit files in `docs/` ONLY when explicitly instructed in the user prompt.
- **Source Code**: Read `.go` source files selectively when inspecting or modifying specific logic.
- **Code Style**: Ensure Go code modifications follow standard Go conventions, maintain package structures, and pass formatting.