/*
 * Copyright (c) 2025 Clement Li. All rights reserved.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/li-clement/Nemesis/internal/scanner"
)

func main() {
	// Parse command line arguments
	zipFile := flag.String("zip", "", "Path to the zip file to analyze")
	outputFile := flag.String("output", "copyright_analysis.txt", "Path to the output file")
	endpoint := flag.String("endpoint", "", "MCP endpoint URL")
	apiKey := flag.String("api-key", "", "MCP API key")
	model := flag.String("model", "gpt-4", "Model to use for analysis")
	flag.Parse()

	if *zipFile == "" {
		fmt.Println("Error: zip file path is required")
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "" {
		fmt.Println("Error: MCP endpoint is required")
		flag.Usage()
		os.Exit(1)
	}

	if *apiKey == "" {
		fmt.Println("Error: MCP API key is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create scanner and MCP service
	s := scanner.NewScanner()
	mcpService, err := scanner.NewMCPService(s, scanner.MCPConfig{
		Model:    *model,
		Endpoint: *endpoint,
		APIKey:   *apiKey,
	})
	if err != nil {
		fmt.Printf("Error creating MCP service: %v\n", err)
		os.Exit(1)
	}

	// Analyze the zip file
	result, err := mcpService.AnalyzeZipFile(context.Background(), *zipFile)
	if err != nil {
		fmt.Printf("Error analyzing zip file: %v\n", err)
		os.Exit(1)
	}

	// Write result to file
	if err := os.WriteFile(*outputFile, []byte(result), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Analysis complete. Results saved to: %s\n", *outputFile)
}
