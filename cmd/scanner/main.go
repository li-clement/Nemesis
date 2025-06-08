/*
 * Copyright (c) 2025 Clement Li. All rights reserved.
 */

package main

import (
	"fmt"
	"os"

	"github.com/li-clement/Nemesis/internal/scanner"
)

func main() {
	// Check command line arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: scanner <scan directory> <output file pattern>")
		fmt.Println("Example: scanner test_files 'copyright_{name}.txt'")
		fmt.Println("Note: {name} will be replaced with subdirectory name")
		os.Exit(1)
	}

	// Create scanner and scan directories
	s := scanner.NewScanner()
	err := s.ScanSubDirectories(os.Args[1], os.Args[2])

	// Handle errors
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All directories scanned successfully!")
}
