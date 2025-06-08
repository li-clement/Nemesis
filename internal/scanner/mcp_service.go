/*
 * Copyright (c) 2025 Clement Li. All rights reserved.
 */

package scanner

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

// MCPClient defines the interface for MCP client operations
type MCPClient interface {
	CallTool(ctx context.Context, tool string, params any) (*mcp.ToolResponse, error)
	GetPrompt(ctx context.Context, tool string, messages any) (*mcp.PromptResponse, error)
}

// MCPService handles the Model Context Protocol integration
type MCPService struct {
	scanner   *Scanner
	mcpClient MCPClient
	model     string
}

// MCPConfig holds the configuration for MCP service
type MCPConfig struct {
	Model    string
	Endpoint string
	APIKey   string
}

// NewMCPService creates a new MCP service instance
func NewMCPService(scanner *Scanner, config MCPConfig) (*MCPService, error) {
	// Create a new MCP client with HTTP transport
	transport := http.NewHTTPClientTransport("/mcp")
	transport.WithBaseURL(config.Endpoint)
	transport.WithHeader("Authorization", "Bearer "+config.APIKey)

	mcpClient := mcp.NewClient(transport)

	return &MCPService{
		scanner:   scanner,
		mcpClient: mcpClient,
		model:     config.Model,
	}, nil
}

// AnalyzeCopyright analyzes copyright information in a zip file
func (s *MCPService) AnalyzeCopyright(zipFile string) (string, error) {
	// Create a temporary directory to extract the zip file
	tempDir, err := os.MkdirTemp("", "nemesis_analysis_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract the zip file
	if err := s.extractZip(zipFile, tempDir); err != nil {
		return "", fmt.Errorf("failed to extract zip file: %v", err)
	}

	// Use Scanner to extract copyright information
	copyrightInfo, err := s.scanner.ScanDirectory(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to scan directory: %v", err)
	}

	// Use MCP to analyze the content
	ctx := context.Background()
	response, err := s.mcpClient.CallTool(ctx, "analyze_copyright", map[string]interface{}{
		"content": copyrightInfo,
	})

	if err != nil {
		return "", fmt.Errorf("failed to analyze content with MCP: %v", err)
	}

	// Extract the analysis from the response
	if response != nil && len(response.Content) > 0 {
		return response.Content[0].TextContent.Text, nil
	}

	return "", fmt.Errorf("no analysis result received from MCP")
}

// AnalyzeZipFile analyzes copyright information in a zip file using MCP
func (m *MCPService) AnalyzeZipFile(ctx context.Context, zipPath string) (string, error) {
	// Create a temporary directory to extract the zip file
	tempDir, err := os.MkdirTemp("", "nemesis_analysis_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract the zip file
	if err := m.extractZip(zipPath, tempDir); err != nil {
		return "", fmt.Errorf("failed to extract zip file: %v", err)
	}

	// Scan the extracted directory for copyright information
	copyrightInfo, err := m.scanner.ScanDirectory(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to scan directory: %v", err)
	}

	// Prepare context for MCP
	messages := []*mcp.PromptMessage{
		mcp.NewPromptMessage(
			mcp.NewTextContent(`You are a copyright analysis expert. Analyze the provided copyright information and provide:
1. A summary of all copyright holders
2. The years covered by the copyrights
3. Any potential conflicts or overlapping claims
4. Recommendations for compliance
Please format your response in a clear, structured manner.`),
			mcp.RoleAssistant,
		),
		mcp.NewPromptMessage(
			mcp.NewTextContent(fmt.Sprintf("Please analyze the following copyright information from a software project:\n\n%s",
				copyrightInfo)),
			mcp.RoleUser,
		),
	}

	// Call MCP for analysis
	response, err := m.mcpClient.GetPrompt(ctx, "analyze_copyright", messages)
	if err != nil {
		return "", fmt.Errorf("failed to get MCP analysis: %v", err)
	}

	// Get the response text from the last message
	var analysisText string
	if len(response.Messages) > 0 {
		lastMessage := response.Messages[len(response.Messages)-1]
		if lastMessage.Content != nil && lastMessage.Content.Type == mcp.ContentTypeText {
			analysisText = lastMessage.Content.TextContent.Text
		}
	}

	// Format and return the result
	return m.formatAnalysisResult(copyrightInfo, analysisText), nil
}

// extractZip extracts a zip file to the specified directory
func (m *MCPService) extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		// Create directory if needed
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Create file
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// Open zip file
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Copy contents
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// formatAnalysisResult formats the analysis result
func (m *MCPService) formatAnalysisResult(copyrightInfo, analysis string) string {
	var result strings.Builder

	result.WriteString("Copyright Analysis Result\n")
	result.WriteString("=======================\n\n")

	result.WriteString("Original Copyright Information:\n")
	result.WriteString("-----------------------------\n")
	result.WriteString(copyrightInfo)
	result.WriteString("\n\n")

	result.WriteString("AI Analysis:\n")
	result.WriteString("-----------\n")
	result.WriteString(analysis)

	return result.String()
}
