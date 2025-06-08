# Nemesis

A Go-based tool designed to scan all text files in a directory, extract copyright information, and save the results to a specified output file.

## Features

- Smart text file detection (automatically skips binary files)
- Intelligent copyright information recognition (supports "copyright", "©", "(c)", and "(C)" identifiers)
- Full text scanning to ensure no copyright information is missed
- Automatic deduplication to avoid duplicate copyright information
- Clear formatted output to file

## Installation

```bash
go get github.com/your-username/copyright-scanner
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
├── NOTICE              # Copyright and attribution notice
└── README.md           # Project documentation
```

## Output Format

Each line in the output file represents a unique copyright notice, for example:

```
Copyright (c) 2024 Example Corp.
Copyright © 2023 Another Company Inc.
Copyright (C) 2022 Open Source Project
```

## License

This project is licensed under the [Apache License 2.0](LICENSE). 