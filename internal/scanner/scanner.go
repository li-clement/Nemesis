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

// Scanner 结构体用于处理版权信息扫描
type Scanner struct {
	// 移除 codeExtensions，因为我们现在扫描所有文本文件
}

// NewScanner 创建一个新的扫描器实例
func NewScanner() *Scanner {
	return &Scanner{}
}

// isTextFile 检查文件是否是文本文件
func (s *Scanner) isTextFile(path string) bool {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// 读取文件的前512字节
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// 检查是否包含空字节（二进制文件的特征）
	if bytes.Contains(buf, []byte{0}) {
		return false
	}

	// 检查文件内容是否是可打印的ASCII字符或常见的Unicode字符
	for _, b := range buf {
		if b < 32 && !isAllowedControlChar(b) {
			return false
		}
	}

	return true
}

// isAllowedControlChar 检查是否是允许的控制字符
func isAllowedControlChar(b byte) bool {
	// 允许的控制字符：换行、回车、制表符
	return b == '\n' || b == '\r' || b == '\t'
}

// cleanLine 清理行中的注释符号和其他标记
func cleanLine(line string) string {
	// 移除开头的注释符号和其他标记
	prefixes := []string{"//", "/*", "*/", "#", "*", "+", "-", "<!--", "-->"}
	trimmed := line

	// 重复清理，直到没有可清理的前缀
	for {
		original := trimmed
		trimmed = strings.TrimSpace(trimmed)

		// 移除所有注释符号，不管它们在哪里
		for _, prefix := range prefixes {
			trimmed = strings.ReplaceAll(trimmed, prefix, " ")
		}

		// 规范化空白字符
		trimmed = strings.Join(strings.Fields(trimmed), " ")

		if original == trimmed {
			break
		}
	}

	return trimmed
}

// normalizeForComparison 标准化字符串以进行比较
func normalizeForComparison(s string) string {
	// 转换为小写
	s = strings.ToLower(s)

	// 移除所有标点符号（包括句点）和特殊字符
	s = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return ' '
		}
		return unicode.ToLower(r)
	}, s)

	// 规范化空白字符
	fields := strings.Fields(s)

	// 移除常见的前缀词和年份
	var cleanFields []string
	for i := 0; i < len(fields); i++ {
		field := fields[i]

		// 跳过常见的前缀词
		if field == "copyright" || field == "c" || field == "by" ||
			field == "corp" || field == "corporation" || field == "inc" ||
			field == "affiliates" || field == "all" || field == "rights" ||
			field == "reserved" || field == "and" || field == "the" ||
			field == "team" || field == "authors" || field == "license" {
			continue
		}

		// 跳过年份（4位数字）
		if len(field) == 4 {
			if _, err := strconv.Atoi(field); err == nil {
				continue
			}
		}

		// 跳过年份范围（例如：2022-2025）
		if i < len(fields)-2 && len(field) == 4 {
			if year1, err1 := strconv.Atoi(field); err1 == nil {
				if fields[i+1] == "-" || fields[i+1] == "to" {
					if year2, err2 := strconv.Atoi(fields[i+2]); err2 == nil {
						if year2 > year1 && year2-year1 <= 100 { // 确保是合理的年份范围
							i += 2 // 跳过分隔符和第二个年份
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

// extractCopyright 从文件中提取copyright信息
func (s *Scanner) extractCopyright(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 设置更大的缓冲区
	reader := bufio.NewReaderSize(file, 1024*1024) // 1MB buffer
	var copyright strings.Builder
	seenCopyrights := make(map[string]bool)

	// 用于存储多行版权信息
	var currentCopyright strings.Builder
	var isCollectingCopyright bool

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		// 去除前后空白
		trimmedLine := strings.TrimSpace(line)

		// 处理空行
		if trimmedLine == "" {
			if isCollectingCopyright {
				// 处理已收集的版权信息
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

		// 跳过可能的代码行和测试相关内容
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
				// 处理已收集的版权信息
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

		// 检查是否包含copyright相关文字，并确保这是一个真实的版权声明
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

			// 开始收集版权信息
			isCollectingCopyright = true
			currentCopyright.WriteString(trimmedLine)
		} else if isCollectingCopyright {
			// 继续收集版权信息
			currentCopyright.WriteString(" " + trimmedLine)
		}

		if err == io.EOF {
			// 处理最后一个版权信息
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

// ScanSubDirectories 扫描指定目录下的所有子目录
func (s *Scanner) ScanSubDirectories(rootDir string, outputPattern string) error {
	// 获取所有子目录
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(rootDir, entry.Name())

			// 生成输出文件名
			outputFile := strings.ReplaceAll(outputPattern, "{name}", entry.Name())
			if !strings.Contains(outputPattern, "{name}") {
				// 如果模式中没有 {name}，在文件名和扩展名之间插入目录名
				ext := filepath.Ext(outputFile)
				base := strings.TrimSuffix(outputFile, ext)
				if strings.HasSuffix(base, "_") {
					base = strings.TrimSuffix(base, "_")
				}
				outputFile = base + "_" + entry.Name() + ext
			}

			// 扫描子目录
			copyrightText, err := s.ScanDirectory(subDir)
			if err != nil {
				return fmt.Errorf("扫描目录 %s 失败: %v", subDir, err)
			}

			// 从 template 文件夹读取 prefix.txt 的内容
			prefixContent := ""
			if prefixBytes, err := os.ReadFile("template/prefix.txt"); err == nil {
				prefixContent = string(prefixBytes)

				// 在 prefix.txt 中查找并替换 Software: 行
				lines := strings.Split(prefixContent, "\n")
				for i, line := range lines {
					if strings.TrimSpace(line) == "Software:" {
						lines[i] = "Software: " + entry.Name()
						break
					}
				}
				prefixContent = strings.Join(lines, "\n")

				// 确保前缀内容以换行符结束
				if !strings.HasSuffix(prefixContent, "\n") {
					prefixContent += "\n"
				}

				// 组合前缀和版权信息
				copyrightText = prefixContent + copyrightText
			}

			// 写入结果
			if err := os.WriteFile(outputFile, []byte(copyrightText), 0644); err != nil {
				return fmt.Errorf("写入文件 %s 失败: %v", outputFile, err)
			}

			fmt.Printf("完成扫描 %s，结果已保存到: %s\n", subDir, outputFile)
		}
	}

	return nil
}

// ScanDirectory 扫描单个目录
func (s *Scanner) ScanDirectory(dir string) (string, error) {
	var result strings.Builder
	seenCopyrights := make(map[string]bool)

	// 首先查找并读取 LICENSE 文件
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

		// 跳过目录和非文本文件
		if info.IsDir() {
			return nil
		}
		if !s.isTextFile(path) {
			return nil
		}

		// 提取copyright信息
		copyright, err := s.extractCopyright(path)
		if err != nil {
			fmt.Printf("处理文件 %s 时出错: %v\n", path, err)
			return nil
		}

		// 如果找到copyright信息，添加到结果中（避免重复）
		if copyright != "" {
			// 分割多行copyright信息
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
		return "", fmt.Errorf("扫描目录时出错: %v", err)
	}

	// 如果找到了 LICENSE 文件，添加到结果末尾
	if licenseContent != "" {
		// 添加一个分隔行
		result.WriteString("\nLicense Text:\n")
		result.WriteString("----------------------------------------\n\n")
		result.WriteString(licenseContent)

		// 确保文件以换行符结束
		if !strings.HasSuffix(licenseContent, "\n") {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}
