# Echogy

A lightweight and efficient SSH reverse proxy tool implemented in Go, featuring a beautiful Terminal User Interface (TUI).

## Features

### Terminal User Interface (TUI)
- Modern and responsive terminal interface
- SSH session management
- Real-time connection status monitoring
- User-friendly interface for managing SSH connections

### Core Features
- SSH reverse proxy functionality
- Multiple concurrent SSH connections support
- TCP port forwarding
- Secure session management
- Built-in logging system

## Quick Start

1. Clone the repository:
```shell
git clone https://github.com/youkale/echogy.git
cd echogy
```

2. Install dependencies:
```shell
go mod download
```

3. Configure your settings in `config.json`:
```json
{
  "addr": ":443",
  "ssh_addr": ":22",
  "domain": "your-domain.com",
  "idle_timeout": 300,
  "key": "YOUR_SSH_KEY"
}
```

4. Build and run:
```shell
make build
./echogy
```

## Project Structure
```
.
├── cmd/           # Command line tools
├── logger/        # Logging framework
├── tui/          # Terminal User Interface components
├── pprof/        # Performance profiling
├── echogy.go     # Core SSH implementation
├── conn.go       # Connection management
├── facade.go     # Facade pattern implementation
├── forward.go    # Port forwarding logic
└── util.go       # Utility functions
```

### Key Components
- SSH Server: Handles SSH connections and session management
- Forward Proxy: Manages TCP port forwarding
- TUI: Provides an interactive terminal interface
- Logger: Structured logging with multiple output formats

## Configuration

### SSH Key Setup
```shell
ssh-keygen -b 2048 -f echogy_rsa
# Copy the private key content to config.json
```

### Domain Configuration
```shell
# DNS A records
A your-domain.com YOUR_SERVER_IP
A *.your-domain.com YOUR_SERVER_IP
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the BSD License - see the LICENSE file for details.