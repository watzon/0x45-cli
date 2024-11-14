# 0x45-cli

<div align="center">
    <img src="https://raw.githubusercontent.com/watzon/0x45-cli/main/.github/0x45.png" alt="0x45 Logo" />
</div>

<div align="center">
    <a href="https://pkg.go.dev/github.com/watzon/0x45-cli">
        <img src="https://pkg.go.dev/badge/github.com/watzon/0x45-cli.svg" alt="Go Reference">
    </a>
    <a href="https://goreportcard.com/report/github.com/watzon/0x45-cli">
        <img src="https://goreportcard.com/badge/github.com/watzon/0x45-cli" alt="Go Report Card">
    </a>
    <a href="https://github.com/watzon/0x45-cli/actions/workflows/test.yml">
        <img src="https://github.com/watzon/0x45-cli/actions/workflows/test.yml/badge.svg" alt="Test Status">
    </a>
    <a href="https://codecov.io/gh/watzon/0x45-cli">
        <img src="https://codecov.io/gh/watzon/0x45-cli/branch/main/graph/badge.svg" alt="Code Coverage">
    </a>
    <a href="https://github.com/watzon/0x45-cli/blob/main/LICENSE">
        <img src="https://img.shields.io/github/license/watzon/0x45-cli" alt="License">
    </a>
</div>

A command-line interface for interacting with the [0x45.st](https://0x45.st) file and URL sharing service.

## Features

- 📤 Upload files and get shareable links
- 🔗 Shorten URLs
- 📋 List your uploaded files and shortened URLs
- 🗑️ Delete uploaded content
- ⚙️ Configurable settings
- 🔑 API key management

## Installation

### From Source

Requires Go 1.18 or later.

```bash
go install github.com/watzon/0x45-cli/0x45@latest
```

The command will be installed as `0x45` in your `$GOPATH/bin` directory.

## Configuration

Before using the CLI, you'll need to configure it with your 0x45.st API key:

```bash
0x45 config set api_key YOUR_API_KEY
```

You can also configure the API URL if you're using a self-hosted instance:

```bash
0x45 config set api_url https://your-instance.com
```

## Usage

### Upload a File

```bash
0x45 upload path/to/file.txt
```

Options:
- `--private`: Make the upload private
- `--expires`: Set expiration time (e.g., "24h", "7d", "1month")

### Shorten a URL

```bash
0x45 shorten https://very-long-url.com
```

Options:
- `--private`: Make the shortened URL private
- `--expires`: Set expiration time

### List Your Content

List all your uploads:
```bash
0x45 list pastes
```

List shortened URLs:
```bash
0x45 list urls
```

Options:
- `--page`: Page number for pagination
- `--limit`: Number of items per page

### Delete Content

```bash
0x45 delete CONTENT_ID
```

### Configuration Management

Get a config value:
```bash
0x45 config get KEY
```

Set a config value:
```bash
0x45 config set KEY VALUE
```

## API Key

To get an API key, visit [0x45.st](https://0x45.st) and request one using:

```bash
0x45 key request --name "You Name" --email "your@email.com"
```

## Development

### Requirements

- Go 1.16 or later
- [golangci-lint](https://golangci-lint.run/) for linting

### Running Tests

```bash
go test -v ./...
```

### Linting

```bash
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Credits

Built by [watzon](https://github.com/watzon) for use with [0x45.st](https://0x45.st)
