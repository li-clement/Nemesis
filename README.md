# Nemesis

A Go-based tool designed to scan all text files in a directory, extract copyright information, and analyze it using Model Context Protocol (MCP).

We also provide documents generated and hosted by **DeepWiki**, see [here](https://deepwiki.com/li-clement/Nemesis)

## Features

- Smart text file detection (automatically skips binary files)
- Intelligent copyright information recognition (supports "copyright", "©", "(c)", and "(C)" identifiers)
- Full text scanning to ensure no copyright information is missed
- Automatic deduplication to avoid duplicate copyright information
- MCP integration for advanced copyright analysis
  - Summarization of copyright holders
  - Analysis of copyright years and durations
  - Detection of potential conflicts
  - Compliance recommendations
- Support for both single files and ZIP archives
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

### Basic Copyright Scanning

```bash
copyright-scanner <scan directory> <output file>
```

Example:
```bash
copyright-scanner . copyright_results.txt
```

### MCP Analysis

To use the MCP analysis features, you'll need to set up your MCP configuration:

```go
config := MCPConfig{
    Model:    "your-model",
    Endpoint: "your-mcp-endpoint",
    APIKey:   "your-api-key",
}

scanner := scanner.NewScanner()
mcpService, err := scanner.NewMCPService(scanner, config)
```

Then you can use the following methods:

1. Analyze a ZIP file:
```go
result, err := mcpService.AnalyzeCopyright("path/to/your.zip")
```

2. Analyze with detailed report:
```go
ctx := context.Background()
result, err := mcpService.AnalyzeZipFile(ctx, "path/to/your.zip")
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

### Basic Scanning Output

Each line in the output file represents a unique copyright notice, for example:

```
Copyright (c) 2024 Example Corp.
Copyright © 2023 Another Company Inc.
Copyright (C) 2022 Open Source Project
```

### MCP Analysis Output

The MCP analysis provides a structured report including:

```
Copyright Analysis Report
------------------------

1. Copyright Holders Summary:
   - Example Corp.
   - Another Company Inc.
   - Open Source Project

2. Copyright Years:
   - Range: 2022-2024
   - Active copyrights: 3

3. Potential Conflicts:
   - No conflicts detected

4. Compliance Recommendations:
   - Ensure all derivative works maintain copyright notices
   - Include license text in distribution
```

## Dependencies

- Go 1.23 or later
- MCP SDK (github.com/metoro-io/mcp-golang)

## Roadmap

The following features are planned for future releases:

### Near-term Goals
- [x] MCP (Model Context Protocol) Integration
  - Implementation of Model Context Protocol for enhanced context management
  - Support for dynamic context switching and handling
  - Improved model interaction and response generation
  - Context-aware processing capabilities
  - Integration with various LLM providers
- [ ] Advanced context manipulation features
- [ ] Performance optimization for large-scale processing

### Future Enhancements
- [ ] Integration with CI/CD pipelines
- [ ] Web interface for easier interaction
- [ ] Support for multiple model providers
- [ ] Custom context definition and management
- [ ] Batch processing capabilities
- [ ] Advanced conflict detection algorithms

## License

This project is licensed under the [Apache License 2.0](LICENSE). 
