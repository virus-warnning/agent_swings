# Directory structure

```
├── Taskfile.yaml                    # 🛠️ Task build/run scripts (build, run, containers, clean)
├── compose/                         # 🐳 Service dependency containers
│   └── docker-compose.yaml          # 🐳 Kroki (Mermaid) container definition
├── docs/                            # 📚 Project documentation
│   └── SYS_ARCH.md                  # 📐 System architecture diagram (Mermaid)
├── src/                             # 💻 Source code
│   └── backend/                     # 🐹 Go backend module
│       ├── go.mod                   # 📦 Go module and dependency declarations
│       ├── go.sum                   # 🔒 Dependency version lock
│       ├── cmd/                     # 🚀 Executable entry point
│       │   └── swings/              # 🎯 swings executable
│       │       └── swings.go        # ▶️ main entry
│       └── internal/                # 🧱 Internal packages
│           ├── commons/             # 🧰 Shared utilities
│           │   └── path_utils.go    # 🧭 Path utility functions
│           └── handlers/            # 📡 Request handlers
│               ├── mcp_handlers.go  # 🤖 MCP protocol handler (2026-07-28, stateless session)
│               └── mcp_types.go     # 📝 MCP data type definitions
└── zaplog.yaml                      # 🪵 zap logger configuration
```
