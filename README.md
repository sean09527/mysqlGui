# MySQL Manager

A modern, cross-platform MySQL/MariaDB database management tool built with Wails, Go, and Vue3.

## Features

- 🔌 **Connection Management**: Manage multiple database connections with SSH tunnel support
- 📊 **Schema Management**: Create, modify, and delete tables with an intuitive interface
- 📝 **Data Management**: View, insert, update, and delete data with filtering and sorting
- 🔍 **SQL Query Editor**: Execute custom SQL queries with syntax highlighting
- 🔄 **Schema Synchronization**: Compare and sync database structures across environments
- 📦 **Import/Export**: Support for SQL, CSV, and JSON formats
- 🔒 **Security**: Encrypted password storage and SQL injection protection

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Wails CLI v2

### Installation

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repository
git clone https://github.com/sean09527/mysqlGui.git
cd mysqlGui

```

### Development

```bash
# Run in development mode with hot reload
make dev

# Or use wails directly
wails dev
```

The application will start with hot reload enabled. Changes to frontend code will be reflected immediately.

## Building

### Build for Current Platform

```bash
# Using Make
make build

# Or using Wails directly
wails build
```

### Build for All Platforms

```bash
# Build for Windows, macOS, and Linux
make build-all

# Or use the build script
./build/build-all.sh
```

### Build for Specific Platform

```bash
# Windows
make build-windows

# macOS (Universal Binary)
make build-macos

# Linux
make build-linux
```

### Create Installers

```bash
# Windows NSIS installer
make package-windows

# macOS DMG
make package-macos

# Linux DEB package
make package-linux
```

## Documentation

- [Build Guide](build/BUILD.md) - Detailed build instructions
- [Packaging Guide](build/PACKAGING.md) - Creating installers and packages
- [Icon Guide](build/ICONS.md) - Managing application icons
- [Release Checklist](build/RELEASE_CHECKLIST.md) - Pre-release checklist

## Project Structure

```
mysql-manager/
├── backend/              # Go backend code
│   ├── internal/         # Internal packages
│   │   ├── connection/   # Connection management
│   │   ├── schema/       # Schema management
│   │   ├── data/         # Data management
│   │   ├── sync/         # Schema synchronization
│   │   └── ...
│   └── *.go              # API bindings
├── frontend/             # Vue3 frontend
│   ├── src/
│   │   ├── components/   # Vue components
│   │   ├── views/        # Page views
│   │   ├── stores/       # Pinia stores
│   │   └── ...
│   └── package.json
├── build/                # Build configuration and assets
│   ├── appicon.png       # Application icon
│   ├── windows/          # Windows-specific assets
│   ├── darwin/           # macOS-specific assets
│   └── *.md              # Build documentation
├── wails.json            # Wails configuration
├── Makefile              # Build automation
└── README.md
```

## Technology Stack

### Backend
- **Go 1.21+**: High-performance backend
- **Wails v2**: Desktop application framework
- **go-sql-driver/mysql**: MySQL database driver
- **SQLite**: Local configuration storage

### Frontend
- **Vue 3**: Progressive JavaScript framework
- **TypeScript**: Type-safe development
- **Element Plus**: UI component library
- **CodeMirror**: SQL editor with syntax highlighting
- **Pinia**: State management

## Database Compatibility

- MySQL 5.7+
- MySQL 8.0+
- MariaDB 10.2+
- MariaDB 10.5+

## Platform Support

- ✅ Windows 10/11
- ✅ macOS 10.15+ (Intel and Apple Silicon)
- ✅ Linux (Ubuntu, Fedora, Arch, etc.)

## Testing

```bash
# Run all tests
make test

# Run backend tests only
make test-backend

# Run frontend tests only
make test-frontend
```

## Version Management

```bash
# Show current version
./build/version.sh show

# Set specific version
./build/version.sh set 1.2.3

# Bump version
./build/version.sh bump major   # 1.0.0 -> 2.0.0
./build/version.sh bump minor   # 1.0.0 -> 1.1.0
./build/version.sh bump patch   # 1.0.0 -> 1.0.1
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/yourusername/mysql-manager/issues)
- 💬 [Discussions](https://github.com/yourusername/mysql-manager/discussions)

## Acknowledgments

- Built with [Wails](https://wails.io/)
- UI components from [Element Plus](https://element-plus.org/)
- Icons from [Heroicons](https://heroicons.com/)

---

Made with ❤️ by the MySQL Manager team
