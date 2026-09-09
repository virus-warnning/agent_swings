# Directory structure

```
├── Taskfile.yaml                    # 🛠️ Task build/run scripts (build, run, containers, clean)
├── compose/                         # 🐳 Service dependency containers
│   └── docker-compose.yaml          # 🐳 Kroki (Mermaid) container definition
├── docs/                            # 📚 Project documentation
│   └── SYS_ARCH.md                  # 📐 System architecture diagram (Mermaid)
├── src/                             # 💻 Source code
│   ├── backend/                     # 🐹 Go backend module
│   │   ├── go.mod                   # 📦 Go module and dependency declarations
│   │   ├── go.sum                   # 🔒 Dependency version lock
│   │   ├── cmd/                     # 🚀 Executable entry point
│   │   │   └── swings/              # 🎯 swings executable
│   │   │       └── swings.go        # ▶️ main entry
│   │   └── internal/                # 🧱 Internal packages
│   │       ├── commons/             # 🧰 Shared utilities
│   │       │   └── path_utils.go    # 🧭 Path utility functions
│   │       └── handlers/            # 📡 Request handlers
│   │           ├── mcp_handlers.go  # 🤖 MCP protocol handler (2026-07-28, stateless session)
│   │           └── mcp_types.go     # 📝 MCP data type definitions
│   └── frontend/                    # ⚡ Nuxt frontend
│       ├── nuxt.config.ts           # ⚙️ Nuxt framework configuration
│       ├── package.json             # 📦 npm dependencies
│       ├── pnpm-workspace.yaml      # 📦 pnpm workspace configuration
│       ├── tsconfig.json            # 📝 TypeScript configuration
│       ├── app/                     # 📱 Nuxt app directory
│       │   ├── app.config.ts        # ⚙️ App runtime configuration
│       │   ├── app.vue              # 🖥️ Root app component
│       │   ├── assets/              # 🎨 Static assets
│       │   │   └── css/
│       │   │       └── main.css     # 🎨 Global stylesheet
│       │   ├── components/          # 🧩 Vue components
│       │   │   ├── AppLogo.vue      # 🖼️ Application logo component
│       │   │   └── TemplateMenu.vue # 📋 Template menu component
│       │   └── pages/               # 📄 Route pages
│       │       └── index.vue        # 🏠 Home page
│       └── public/                  # 🌐 Public static files
│           └── favicon.ico          # 🖼️ Site favicon
└── zaplog.yaml                      # 🪵 zap logger configuration
```
