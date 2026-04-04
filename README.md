# Rikkahub2Kelivo 备份转换工具

在 [Kelivo](https://github.com/nicepkg/kelivo) 和 [RikkaHub](https://github.com/rikka/rikkahub) 两个 AI 聊天应用之间互相转换备份文件。

## 功能

- ✅ 提供商/模型配置转换
- ✅ 助手配置转换
- ✅ MCP 服务器配置转换
- ✅ 世界书 / 快捷短语转换
- ✅ 聊天会话 + 聊天记录转换（含多版本分支）
- ✅ 双向转换：Kelivo ↔ RikkaHub

## 使用方式

### 交互式脚本（推荐）

```bash
./run.sh
```

脚本会显示彩色菜单，引导你完成转换操作。

### 命令行

```bash
# 查看备份信息
./backup-converter info -i kelivo_backup.zip
./backup-converter info -i rikkahub_backup.zip

# 转换
./backup-converter kelivo2rikkahub -i kelivo_backup.zip -o rikkahub_output.zip
./backup-converter rikkahub2kelivo -i rikkahub_backup.zip -o kelivo_output.zip
```

## 从源码构建

```bash
cd converter
go build -o backup-converter .
```

需要 Go 1.21+ 和 CGO（用于 SQLite 支持）。

## 项目结构

```
├── main.go                          # CLI 入口
├── models/
│   ├── kelivo.go                    # Kelivo 数据模型
│   └── rikkahub.go                  # RikkaHub 数据模型
├── parser/
│   ├── kelivo_parser.go             # Kelivo 备份解析 (zip)
│   └── rikkahub_parser.go           # RikkaHub 备份解析 (zip + SQLite)
├── converter/
│   ├── kelivo_to_rikkahub.go        # Kelivo → RikkaHub 转换
│   └── rikkahub_to_kelivo.go        # RikkaHub → Kelivo 转换
└── writer/
    ├── kelivo_writer.go             # 写出 Kelivo 备份 zip
    └── rikkahub_writer.go           # 写出 RikkaHub 备份 zip (含 SQLite)
```

## License

Apache License 2.0
