# Kelivo ↔ RikkaHub 备份转换工具

在 [Kelivo](https://github.com/nicepkg/kelivo) 和 [RikkaHub](https://github.com/rikka/rikkahub) 两个 AI 聊天应用之间互相转换备份文件。

## 功能

- 提供商/模型配置转换
- 助手配置转换
- MCP 服务器配置转换
- 世界书/快捷短语转换
- 聊天会话+记录转换（含多版本分支）
- 双向转换：Kelivo ↔ RikkaHub

## 下载

从 [Releases](https://github.com/jwbb903/Rikkahub2Kelivo/releases) 下载预编译的二进制文件。

## 快速开始

### 交互式脚本（推荐）

```bash
./run.sh
```

### 命令行

```bash
# Kelivo → RikkaHub
./backup-converter k2r -i kelivo_backup.zip -o output.zip

# RikkaHub → Kelivo
./backup-converter r2k -i rikkahub_backup.zip -o output.zip

# Web 服务器
./backup-converter serve -p 8080
```

## 从源码构建

```bash
go build -o backup-converter .
```

需要 Go 1.21+ 和 CGO（用于 SQLite 支持）。

## 项目结构

```
├── main.go              # CLI 入口
├── server.go            # Web 服务器
├── web/                 # 前端页面
├── models/              # 数据模型
├── parser/              # 备份解析
├── converter/           # 转换逻辑
└── writer/              # 备份导出
```

## License

Apache License 2.0
