/*
 * Copyright (c) 2025 Clement Li. All rights reserved.
 */

package main

import (
	"fmt"
	"os"

	"nemesis/internal/scanner"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("使用方法: scanner <扫描目录> <输出文件模式>")
		fmt.Println("例如: scanner test_files 'copyright_{name}.txt'")
		fmt.Println("注意: {name} 将被替换为子目录名")
		os.Exit(1)
	}

	scanDir := os.Args[1]
	outputPattern := os.Args[2]

	scanner := scanner.NewScanner()
	if err := scanner.ScanSubDirectories(scanDir, outputPattern); err != nil {
		fmt.Printf("扫描出错: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("所有目录扫描完成！")
}
