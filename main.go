package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/converter/backup-converter/converter"
	"github.com/converter/backup-converter/parser"
	"github.com/converter/backup-converter/writer"
)

const version = "1.3.0"

func printUsage() {
	fmt.Printf(`备份转换工具 v%s
在 Kelivo 和 RikkaHub 两个 AI 聊天应用之间互相转换备份文件。

用法:
  backup-converter <命令> [选项]

命令:
  k2r, kelivo2rikkahub  将 Kelivo 备份转换为 RikkaHub 格式
  r2k, rikkahub2kelivo  将 RikkaHub 备份转换为 Kelivo 格式
  i, info               查看备份文件的详细信息
  serve                 启动 Web 服务器（上传、转换、查看）

选项:
  -i, --input   <文件>    输入备份 zip 文件（必填，serve 除外）
  -o, --output  <文件>    输出备份 zip 文件（可选，不填自动生成）
  -p, --port    <端口>    Web 服务器端口（默认 8080）
  -h, --help              显示帮助信息
  -v, --version           显示版本信息

示例:
  backup-converter k2r -i kelivo_backup.zip -o rikkahub_backup.zip
  backup-converter r2k -i rikkahub_backup.zip
  backup-converter i kelivo_backup.zip
  backup-converter serve -p 3000
`, version)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "-h", "--help", "help":
		printUsage()
		return
	case "-v", "--version", "version":
		fmt.Printf("备份转换工具 v%s\n", version)
		return
	case "k2r", "kelivo2rikkahub":
		handleKelivo2RikkaHub(os.Args[2:])
	case "r2k", "rikkahub2kelivo":
		handleRikkaHub2Kelivo(os.Args[2:])
	case "i", "info":
		handleInfo(os.Args[2:])
	case "serve":
		handleServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "错误：未知命令 '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func parseArgs(args []string) (input, output string, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i", "--input":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("参数 %s 需要一个值", args[i])
			}
			i++
			input = args[i]
		case "-o", "--output":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("参数 %s 需要一个值", args[i])
			}
			i++
			output = args[i]
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if input == "" {
				input = args[i]
			} else if output == "" {
				output = args[i]
			}
		}
	}
	return input, output, nil
}

func parseServeArgs(args []string) string {
	port := "8080"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p", "--port":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "错误：参数 %s 需要一个值\n", args[i])
				os.Exit(1)
			}
			i++
			port = args[i]
		default:
			port = args[i]
		}
	}
	return port
}

func handleServe(args []string) {
	port := parseServeArgs(args)
	StartServer(port)
}

func handleKelivo2RikkaHub(args []string) {
	input, output, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(1)
	}

	if input == "" {
		fmt.Fprintf(os.Stderr, "错误：必须指定输入文件\n")
		printUsage()
		os.Exit(1)
	}

	if output == "" {
		output = generateOutputName(input, "rikkahub")
	}

	fmt.Printf("正在将 Kelivo 备份转换为 RikkaHub 格式...\n")
	fmt.Printf("  输入文件：%s\n", input)
	fmt.Printf("  输出文件：%s\n", output)

	fmt.Printf("\n解析 Kelivo 备份中...\n")
	kelivoBackup, err := parser.ParseKelivoBackup(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析失败：%v\n", err)
		os.Exit(1)
	}

	if kelivoBackup.Settings != nil {
		fmt.Printf("  - 提供商数量：%d\n", len(kelivoBackup.Settings.ProviderConfigs))
		fmt.Printf("  - 助手数量：%d\n", len(kelivoBackup.Settings.Assistants))
		fmt.Printf("  - MCP 服务器：%d\n", len(kelivoBackup.Settings.MCPServers))
		fmt.Printf("  - 世界书数量：%d\n", len(kelivoBackup.Settings.WorldBooks))
		fmt.Printf("  - 快捷短语：%d\n", len(kelivoBackup.Settings.QuickPhrases))
	}
	if kelivoBackup.Chats != nil {
		fmt.Printf("  - 聊天会话：%d\n", len(kelivoBackup.Chats.Conversations))
		fmt.Printf("  - 聊天消息：%d\n", len(kelivoBackup.Chats.Messages))
	}

	fmt.Printf("\n正在转换...\n")
	rikkaBackup, err := converter.ConvertKelivoToRikkaHub(kelivoBackup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "转换失败：%v\n", err)
		os.Exit(1)
	}

	if rikkaBackup.Settings != nil {
		fmt.Printf("  - 提供商数量：%d\n", len(rikkaBackup.Settings.Providers))
		fmt.Printf("  - 助手数量：%d\n", len(rikkaBackup.Settings.Assistants))
		fmt.Printf("  - MCP 服务器：%d\n", len(rikkaBackup.Settings.MCPServers))
	}
	fmt.Printf("  - 聊天会话：%d\n", len(rikkaBackup.Conversations))
	fmt.Printf("  - 消息节点：%d\n", len(rikkaBackup.MessageNodes))

	fmt.Printf("\n正在写入 RikkaHub 备份...\n")
	if err := writer.WriteRikkaHubBackup(rikkaBackup, output); err != nil {
		fmt.Fprintf(os.Stderr, "写入失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n转换完成！输出文件已保存到：%s\n", output)
}

func handleRikkaHub2Kelivo(args []string) {
	input, output, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(1)
	}

	if input == "" {
		fmt.Fprintf(os.Stderr, "错误：必须指定输入文件\n")
		printUsage()
		os.Exit(1)
	}

	if output == "" {
		output = generateOutputName(input, "kelivo")
	}

	fmt.Printf("正在将 RikkaHub 备份转换为 Kelivo 格式...\n")
	fmt.Printf("  输入文件：%s\n", input)
	fmt.Printf("  输出文件：%s\n", output)

	fmt.Printf("\n解析 RikkaHub 备份中...\n")
	rikkaBackup, err := parser.ParseRikkaHubBackup(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析失败：%v\n", err)
		os.Exit(1)
	}

	if rikkaBackup.Settings != nil {
		fmt.Printf("  - 提供商数量：%d\n", len(rikkaBackup.Settings.Providers))
		fmt.Printf("  - 助手数量：%d\n", len(rikkaBackup.Settings.Assistants))
		fmt.Printf("  - MCP 服务器：%d\n", len(rikkaBackup.Settings.MCPServers))
		fmt.Printf("  - 世界书数量：%d\n", len(rikkaBackup.Settings.Lorebooks))
		fmt.Printf("  - 快捷消息：%d\n", len(rikkaBackup.Settings.QuickMessages))
	}
	fmt.Printf("  - 聊天会话：%d\n", len(rikkaBackup.Conversations))
	fmt.Printf("  - 消息节点：%d\n", len(rikkaBackup.MessageNodes))
	fmt.Printf("  - 附件文件：%d\n", len(rikkaBackup.UploadFiles))

	fmt.Printf("\n正在转换...\n")
	kelivoBackup, err := converter.ConvertRikkaHubToKelivo(rikkaBackup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "转换失败：%v\n", err)
		os.Exit(1)
	}

	if kelivoBackup.Settings != nil {
		fmt.Printf("  - 提供商数量：%d\n", len(kelivoBackup.Settings.ProviderConfigs))
		fmt.Printf("  - 助手数量：%d\n", len(kelivoBackup.Settings.Assistants))
		fmt.Printf("  - MCP 服务器：%d\n", len(kelivoBackup.Settings.MCPServers))
	}
	if kelivoBackup.Chats != nil {
		fmt.Printf("  - 聊天会话：%d\n", len(kelivoBackup.Chats.Conversations))
		fmt.Printf("  - 聊天消息：%d\n", len(kelivoBackup.Chats.Messages))
	}

	fmt.Printf("\n正在写入 Kelivo 备份...\n")
	if err := writer.WriteKelivoBackup(kelivoBackup, output); err != nil {
		fmt.Fprintf(os.Stderr, "写入失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n转换完成！输出文件已保存到：%s\n", output)
}

func handleInfo(args []string) {
	input, _, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(1)
	}

	if input == "" {
		fmt.Fprintf(os.Stderr, "错误：必须指定输入文件\n")
		printUsage()
		os.Exit(1)
	}

	appType := detectBackupType(input)
	fmt.Printf("备份文件：%s\n", input)
	fmt.Printf("识别格式：%s\n\n", formatDisplayName(appType))

	switch appType {
	case "kelivo":
		backup, err := parser.ParseKelivoBackup(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析失败：%v\n", err)
			os.Exit(1)
		}
		printKelivoInfo(backup)
	case "rikkahub":
		backup, err := parser.ParseRikkaHubBackup(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析失败：%v\n", err)
			os.Exit(1)
		}
		printRikkaHubInfo(backup)
	default:
		fmt.Fprintf(os.Stderr, "未知的备份格式，无法识别\n")
		os.Exit(1)
	}
}

func formatDisplayName(t string) string {
	switch t {
	case "kelivo":
		return "Kelivo"
	case "rikkahub":
		return "RikkaHub"
	default:
		return "未知格式"
	}
}

func detectBackupType(zipPath string) string {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "unknown"
	}
	defer r.Close()

	hasChatsJSON := false
	hasRikkaHubDB := false
	for _, f := range r.File {
		switch f.Name {
		case "chats.json":
			hasChatsJSON = true
		case "rikka_hub.db":
			hasRikkaHubDB = true
		}
	}

	if hasRikkaHubDB {
		return "rikkahub"
	}
	if hasChatsJSON {
		return "kelivo"
	}

	return "unknown"
}

func detectBackupTypeFromBytes(data []byte) string {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "unknown"
	}
	hasChatsJSON := false
	hasRikkaHubDB := false
	for _, f := range r.File {
		switch f.Name {
		case "chats.json":
			hasChatsJSON = true
		case "rikka_hub.db":
			hasRikkaHubDB = true
		}
	}
	if hasRikkaHubDB {
		return "rikkahub"
	}
	if hasChatsJSON {
		return "kelivo"
	}
	return "unknown"
}

func printKelivoInfo(backup *parser.KelivoBackupAlias) {
	fmt.Println("=== Kelivo 备份信息 ===")
	if backup.Settings != nil {
		s := backup.Settings
		fmt.Printf("\n【配置设置】\n")
		fmt.Printf("  提供商数量：%d\n", len(s.ProviderConfigs))
		for id, p := range s.ProviderConfigs {
			fmt.Printf("    - %s (%s)：%d 个模型\n", p.Name, id, len(p.Models))
		}
		fmt.Printf("  助手数量：%d\n", len(s.Assistants))
		for _, a := range s.Assistants {
			fmt.Printf("    - %s (ID: %s)\n", a.Name, a.ID)
		}
		fmt.Printf("  MCP 服务器：%d\n", len(s.MCPServers))
		for _, m := range s.MCPServers {
			fmt.Printf("    - %s（传输方式：%s）\n", m.Name, m.Transport)
		}
		fmt.Printf("  世界书：%d\n", len(s.WorldBooks))
		fmt.Printf("  快捷短语：%d\n", len(s.QuickPhrases))
	}

	if backup.Chats != nil {
		c := backup.Chats
		fmt.Printf("\n【聊天记录】\n")
		fmt.Printf("  会话数量：%d\n", len(c.Conversations))
		fmt.Printf("  消息总数：%d\n", len(c.Messages))
		if len(c.Conversations) > 0 {
			fmt.Printf("\n  最近会话：\n")
			count := 5
			if len(c.Conversations) < count {
				count = len(c.Conversations)
			}
			for i := 0; i < count; i++ {
				conv := c.Conversations[i]
				fmt.Printf("    - %s（%d 条消息）\n", conv.Title, len(conv.MessageIDs))
			}
		}
	}
}

func printRikkaHubInfo(backup *parser.RikkaHubBackupAlias) {
	fmt.Println("=== RikkaHub 备份信息 ===")
	if backup.Settings != nil {
		s := backup.Settings
		fmt.Printf("\n【配置设置】\n")
		fmt.Printf("  提供商数量：%d\n", len(s.Providers))
		for _, p := range s.Providers {
			fmt.Printf("    - %s (%s)：%d 个模型\n", p.Name, p.ID, len(p.Models))
		}
		fmt.Printf("  助手数量：%d\n", len(s.Assistants))
		for _, a := range s.Assistants {
			fmt.Printf("    - %s (ID: %s)\n", a.Name, a.ID)
		}
		fmt.Printf("  MCP 服务器：%d\n", len(s.MCPServers))
		for _, m := range s.MCPServers {
			fmt.Printf("    - %s（类型：%s）\n", m.CommonOptions.Name, m.Type)
		}
		fmt.Printf("  世界书：%d\n", len(s.Lorebooks))
		fmt.Printf("  快捷消息：%d\n", len(s.QuickMessages))
	}

	fmt.Printf("\n【数据库】\n")
	fmt.Printf("  聊天会话：%d\n", len(backup.Conversations))
	fmt.Printf("  消息节点：%d\n", len(backup.MessageNodes))
	fmt.Printf("  管理文件：%d\n", len(backup.ManagedFiles))
	fmt.Printf("  附件文件：%d\n", len(backup.UploadFiles))

	if len(backup.Conversations) > 0 {
		fmt.Printf("\n  最近会话：\n")
		count := 5
		if len(backup.Conversations) < count {
			count = len(backup.Conversations)
		}
		for i := 0; i < count; i++ {
			conv := backup.Conversations[i]
			t := time.UnixMilli(conv.CreateAt).Format("2006-01-02")
			fmt.Printf("    - [%s] %s\n", t, conv.Title)
		}
	}
}

func generateOutputName(input, targetFormat string) string {
	dir := filepath.Dir(input)
	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(dir, fmt.Sprintf("%s_converted_%s.zip", targetFormat, timestamp))
}
