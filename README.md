# WebSSH

A fast reverse proxy tool based on the SSH protocol, built in Go.

## Features

### Terminal User Interface (TUI)
- Beautiful and responsive terminal interface using Bubbletea
- Dynamic tab management for multiple SSH sessions
- Adaptive color themes (light/dark mode support)
- Mouse and window resize support
- Request tracking and management

### Performance & Security
- Fast and no client installation required (*nix like)
- Secure random ID generation for sessions
- Cryptographically secure token generation


### SSH Capabilities
- Multiple concurrent SSH connections
- Secure key-based authentication
- Port forwarding support
- Session management

## Quick Start

1. Clone the repository:
```shell
git clone https://github.com/youkale/webssh.git
cd webssh
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
  "domain": "webs.sh",
  "idle_timeout": 300,
  "key": "YOUR_SSH_KEY"
}
```

4. Build and run:
```shell
make build
./webssh
```

## Configuration

### SSH Key Generation
```shell
ssh-keygen -b 2048 -f webs.sh_rsa
cat webs.sh_rsa # Copy to config.json key field
```

### DNS Configuration
```shell
A webs.sh YOUR_SERVER_IP
A *.webs.sh YOUR_SERVER_IP
```

### Logging Configuration
The application uses structured logging with support for:
- Console output with color-coded levels
- File output with rotation
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)

## Development

### Project Structure
```
.
├── cmd/           # Command line tools
├── logger/        # Logging framework
├── tui/          # Terminal User Interface
│   ├── tabs.go   # Tab management
│   └── tui.go    # Main TUI implementation
├── util.go       # Utility functions
└── webssh.go     # Core SSH implementation
```

### Key Dependencies
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Zerolog](https://github.com/rs/zerolog) - Zero-allocation logging
- [Bubblezone](https://github.com/charmbracelet/bubblezone) - Mouse zones

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License
[BSD](LICENSE)

## Acknowledgments
- [Charm](https://charm.sh/) for the amazing TUI tools