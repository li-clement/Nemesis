# Nemesis

A Go-based tool designed to scan all text files in a directory, extract copyright information, and save the results to a specified output file.

## Features

- Smart text file detection (automatically skips binary files)
- Intelligent copyright information recognition (supports "copyright", "©", "(c)", and "(C)" identifiers)
- Full text scanning to ensure no copyright information is missed
- Automatic deduplication to avoid duplicate copyright information
- Clear formatted output to file

## Installation

You can install Nemesis in one of the following ways:

### Option 1: Clone the repository

```bash
git clone https://github.com/li-clement/Nemesis.git
cd Nemesis
go build ./cmd/scanner
```

### Option 2: Using go get (requires the package to be published)

```bash
go get github.com/li-clement/Nemesis
```

## Usage

```bash
copyright-scanner <scan directory> <output file>
```

Example:
```bash
copyright-scanner . copyright_results.txt
```

## Project Structure

```
.
├── cmd/
│   └── scanner/          # Copyright scanner CLI tool
├── internal/
│   └── scanner/          # Core implementation of copyright scanner
├── go.mod               # Go module definition
├── LICENSE             # Apache 2.0 License
└── README.md           # Project documentation
```

## Output Format

Each line in the output file represents a unique copyright notice, for example:

```
Copyright (c) 2024 Example Corp.
Copyright © 2023 Another Company Inc.
Copyright (C) 2022 Open Source Project
```

## Roadmap

The following features are planned for future releases:

### Near-term Goals
- [ ] MCP (Model Context Protocol) Integration
  - Implementation of Model Context Protocol for enhanced context management
  - Support for dynamic context switching and handling
  - Improved model interaction and response generation
  - Context-aware processing capabilities
  - Integration with various LLM providers

### Future Enhancements
- [ ] Advanced context manipulation features
- [ ] Performance optimization for large-scale processing
- [ ] Integration with CI/CD pipelines
- [ ] Web interface for easier interaction
- [ ] Support for multiple model providers
- [ ] Custom context definition and management

## License

This project is licensed under the [Apache License 2.0](LICENSE). 