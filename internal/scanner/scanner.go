/*
 * Copyright (c) 2025 Clement Li. All rights reserved.
 */

package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Scanner is a struct for handling copyright information scanning
type Scanner struct {
	// Removed codeExtensions as we now scan all text files
}

// NewScanner creates a new scanner instance
func NewScanner() *Scanner {
	return &Scanner{}
}

// isTextFile checks if a file is a text file
func (s *Scanner) isTextFile(path string) bool {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read the first 512 bytes of the file
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// Check if it contains null bytes (characteristic of binary files)
	if bytes.Contains(buf, []byte{0}) {
		return false
	}

	// Check if the file content consists of printable ASCII characters or common Unicode characters
	for _, b := range buf {
		if b < 32 && !isAllowedControlChar(b) {
			return false
		}
	}

	return true
}

// isAllowedControlChar checks if a character is allowed as a control character
func isAllowedControlChar(b byte) bool {
	// Allowed control characters: newline, carriage return, tab
	return b == '\n' || b == '\r' || b == '\t'
}

// cleanLine cleans up comments and other markings in a line
func cleanLine(line string) string {
	// Remove leading comment markings and other markings
	prefixes := []string{"//", "/*", "*/", "#", "*", "+", "-", "<!--", "-->"}
	trimmed := line

	// Repeat cleaning until no more prefixes can be removed
	for {
		original := trimmed
		trimmed = strings.TrimSpace(trimmed)

		// Remove all comment markings, no matter where they are
		for _, prefix := range prefixes {
			trimmed = strings.ReplaceAll(trimmed, prefix, " ")
		}

		// Normalize whitespace characters
		trimmed = strings.Join(strings.Fields(trimmed), " ")

		if original == trimmed {
			break
		}
	}

	return trimmed
}

// normalizeForComparison normalizes a string for comparison
func normalizeForComparison(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove all punctuation (including periods) and special characters
	s = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return ' '
		}
		return unicode.ToLower(r)
	}, s)

	// Normalize whitespace characters
	fields := strings.Fields(s)

	// Remove common prefixes and years
	var cleanFields []string
	for i := 0; i < len(fields); i++ {
		field := fields[i]

		// Skip common prefixes
		if field == "copyright" || field == "c" || field == "by" ||
			field == "corp" || field == "corporation" || field == "inc" ||
			field == "affiliates" || field == "all" || field == "rights" ||
			field == "reserved" || field == "and" || field == "the" ||
			field == "team" || field == "authors" || field == "license" {
			continue
		}

		// Skip years (4-digit numbers)
		if len(field) == 4 {
			if _, err := strconv.Atoi(field); err == nil {
				continue
			}
		}

		// Skip year ranges (e.g., 2022-2025)
		if i < len(fields)-2 && len(field) == 4 {
			if year1, err1 := strconv.Atoi(field); err1 == nil {
				if fields[i+1] == "-" || fields[i+1] == "to" {
					if year2, err2 := strconv.Atoi(fields[i+2]); err2 == nil {
						if year2 > year1 && year2-year1 <= 100 { // Ensure it's a reasonable year range
							i += 2 // Skip over the separator and the second year
							continue
						}
					}
				}
			}
		}

		cleanFields = append(cleanFields, field)
	}

	return strings.Join(cleanFields, " ")
}

// extractCopyright extracts copyright information from a file
func (s *Scanner) extractCopyright(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Set a larger buffer
	reader := bufio.NewReaderSize(file, 1024*1024) // 1MB buffer
	var copyright strings.Builder
	seenCopyrights := make(map[string]bool)

	// For storing multi-line copyright information
	var currentCopyright strings.Builder
	var isCollectingCopyright bool

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		// Remove leading and trailing whitespace
		trimmedLine := strings.TrimSpace(line)

		// Handle empty lines
		if trimmedLine == "" {
			if isCollectingCopyright {
				// Handle collected copyright information
				if currentCopyright.Len() > 0 {
					cleanedCopyright := cleanLine(currentCopyright.String())
					normalizedCopyright := normalizeForComparison(cleanedCopyright)
					if !seenCopyrights[normalizedCopyright] {
						seenCopyrights[normalizedCopyright] = true
						copyright.WriteString(cleanedCopyright + "\n")
					}
					currentCopyright.Reset()
				}
				isCollectingCopyright = false
			}
			if err == io.EOF {
				break
			}
			continue
		}

		// Skip possible code lines and test-related content
		lowercaseLine := strings.ToLower(trimmedLine)
		if strings.Contains(lowercaseLine, "func ") ||
			strings.Contains(lowercaseLine, "type ") ||
			strings.Contains(lowercaseLine, "var ") ||
			strings.Contains(lowercaseLine, "const ") ||
			strings.Contains(lowercaseLine, "package ") ||
			strings.Contains(lowercaseLine, "import ") ||
			strings.Contains(lowercaseLine, "return ") ||
			strings.Contains(lowercaseLine, ":=") ||
			strings.Contains(lowercaseLine, "if ") ||
			strings.Contains(lowercaseLine, "test") ||
			strings.Contains(lowercaseLine, "echo") ||
			strings.Contains(lowercaseLine, "find_") ||
			strings.Contains(lowercaseLine, "append") ||
			strings.Contains(lowercaseLine, "error:") ||
			strings.Contains(lowercaseLine, "grep") ||
			strings.Contains(lowercaseLine, "egrep") ||
			strings.Contains(lowercaseLine, "while ") ||
			strings.Contains(lowercaseLine, "read ") ||
			strings.Contains(lowercaseLine, "|") ||
			strings.Contains(lowercaseLine, "grant of") ||
			strings.Contains(lowercaseLine, "license") ||
			strings.Contains(lowercaseLine, "permission") ||
			strings.Contains(lowercaseLine, "permitted") ||
			strings.Contains(lowercaseLine, "distribute") ||
			strings.Contains(lowercaseLine, "notice") ||
			strings.Contains(lowercaseLine, "provided") ||
			strings.Contains(lowercaseLine, "conditions") ||
			strings.Contains(lowercaseLine, "subject to") ||
			strings.Contains(lowercaseLine, "you may") ||
			strings.Contains(lowercaseLine, "you must") ||
			strings.Contains(lowercaseLine, "shall") ||
			strings.Contains(lowercaseLine, "retain") ||
			strings.Contains(lowercaseLine, "reproduce") {
			if isCollectingCopyright {
				// Handle collected copyright information
				if currentCopyright.Len() > 0 {
					cleanedCopyright := cleanLine(currentCopyright.String())
					normalizedCopyright := normalizeForComparison(cleanedCopyright)
					if !seenCopyrights[normalizedCopyright] {
						seenCopyrights[normalizedCopyright] = true
						copyright.WriteString(cleanedCopyright + "\n")
					}
					currentCopyright.Reset()
				}
				isCollectingCopyright = false
			}
			if err == io.EOF {
				break
			}
			continue
		}

		// Check if it contains copyright-related text and ensure it's a real copyright statement
		if (strings.Contains(lowercaseLine, "copyright") ||
			strings.Contains(lowercaseLine, "©") ||
			strings.Contains(lowercaseLine, "(c)") ||
			strings.Contains(trimmedLine, "(C)")) &&
			!strings.Contains(lowercaseLine, "copyrightadder") &&
			!strings.Contains(lowercaseLine, "copyrighttext") &&
			!strings.Contains(lowercaseLine, "addcopyright") &&
			!strings.Contains(lowercaseLine, "extractcopyright") &&
			!strings.Contains(lowercaseLine, "hascopyright") &&
			!strings.Contains(lowercaseLine, "copyright.sh") &&
			!strings.Contains(lowercaseLine, "copyright notice") &&
			!strings.Contains(lowercaseLine, "copyright owner") &&
			!strings.Contains(lowercaseLine, "copyright holder") &&
			!strings.Contains(lowercaseLine, "above copyright") &&
			!strings.Contains(lowercaseLine, "retain") &&
			!strings.Contains(lowercaseLine, "reproduce") {

			// Start collecting copyright information
			isCollectingCopyright = true
			currentCopyright.WriteString(trimmedLine)
		} else if isCollectingCopyright {
			// Continue collecting copyright information
			currentCopyright.WriteString(" " + trimmedLine)
		}

		if err == io.EOF {
			// Handle last copyright information
			if isCollectingCopyright && currentCopyright.Len() > 0 {
				cleanedCopyright := cleanLine(currentCopyright.String())
				normalizedCopyright := normalizeForComparison(cleanedCopyright)
				if !seenCopyrights[normalizedCopyright] {
					seenCopyrights[normalizedCopyright] = true
					copyright.WriteString(cleanedCopyright + "\n")
				}
			}
			break
		}
	}

	return copyright.String(), nil
}

// ScanSubDirectories scans all subdirectories under a specified directory
func (s *Scanner) ScanSubDirectories(rootDir string, outputPattern string) error {
	// Get all subdirectories
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(rootDir, entry.Name())

			// Generate output file name
			outputFile := strings.ReplaceAll(outputPattern, "{name}", entry.Name())
			if !strings.Contains(outputPattern, "{name}") {
				// If pattern does not contain {name}, insert directory name between file name and extension
				ext := filepath.Ext(outputFile)
				base := strings.TrimSuffix(outputFile, ext)
				base = strings.TrimSuffix(base, "_")
				outputFile = base + "_" + entry.Name() + ext
			}

			// Scan subdirectory
			copyrightText, err := s.ScanDirectory(subDir)
			if err != nil {
				return fmt.Errorf("failed to scan directory %s: %v", subDir, err)
			}

			// Read prefix.txt content from template folder
			prefixContent := ""
			if prefixBytes, err := os.ReadFile("template/prefix.txt"); err == nil {
				prefixContent = string(prefixBytes)

				// Find and replace Software: line in prefix.txt
				lines := strings.Split(prefixContent, "\n")
				for i, line := range lines {
					if strings.TrimSpace(line) == "Software:" {
						lines[i] = "Software: " + entry.Name()
						break
					}
				}
				prefixContent = strings.Join(lines, "\n")

				// Ensure prefix content ends with a newline
				if !strings.HasSuffix(prefixContent, "\n") {
					prefixContent += "\n"
				}

				// Combine prefix and copyright information
				copyrightText = prefixContent + copyrightText
			}

			// Write result
			if err := os.WriteFile(outputFile, []byte(copyrightText), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %v", outputFile, err)
			}

			fmt.Printf("Completed scanning %s, result saved to: %s\n", subDir, outputFile)
		}
	}

	return nil
}

// ScanDirectory scans a single directory
func (s *Scanner) ScanDirectory(dir string) (string, error) {
	var result strings.Builder
	seenCopyrights := make(map[string]bool)

	// First find and read LICENSE file
	var licenseContent string
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "license", "license.txt", "license.md"}
	for _, licenseFile := range licenseFiles {
		content, err := os.ReadFile(filepath.Join(dir, licenseFile))
		if err == nil {
			licenseContent = string(content)
			break
		}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-text files
		if info.IsDir() {
			return nil
		}
		if !s.isTextFile(path) {
			return nil
		}

		// Extract copyright information
		copyright, err := s.extractCopyright(path)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", path, err)
			return nil
		}

		// If copyright information is found, add to result (avoid duplicates)
		if copyright != "" {
			// Split multi-line copyright information
			copyrights := strings.Split(copyright, "\n")
			for _, c := range copyrights {
				if c != "" && !seenCopyrights[c] {
					seenCopyrights[c] = true
					result.WriteString(c + "\n")
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("Error scanning directory: %v", err)
	}

	// If LICENSE file is found, add to result at the end
	if licenseContent != "" {
		// Add a separator line
		result.WriteString("\nLicense Text:\n")
		result.WriteString("----------------------------------------\n\n")
		result.WriteString(licenseContent)

		// Ensure file ends with a newline
		if !strings.HasSuffix(licenseContent, "\n") {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}
