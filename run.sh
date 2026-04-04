#!/bin/bash
# ============================================================
# 备份转换工具 - 方便使用的启动脚本
# 在 Kelivo 和 RikkaHub 两个 AI 聊天应用之间互相转换备份
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # 无颜色

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONVERTER="$SCRIPT_DIR/backup-converter"

# 检查二进制文件是否存在
check_binary() {
    if [ ! -f "$CONVERTER" ]; then
        echo -e "${RED}错误：找不到转换工具 $CONVERTER${NC}"
        echo "请先编译：cd $SCRIPT_DIR && go build -o backup-converter ."
        exit 1
    fi
}

# 显示标题
show_title() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}        ${BOLD}备份转换工具${NC}  ${YELLOW}(Kelivo ⇄ RikkaHub)${NC}        ${CYAN}║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 显示主菜单
show_menu() {
    echo -e "  ${BOLD}请选择操作：${NC}"
    echo ""
    echo -e "    ${GREEN}1)${NC}  ${BOLD}Kelivo → RikkaHub${NC}    将 Kelivo 备份转换为 RikkaHub 格式"
    echo -e "    ${GREEN}2)${NC}  ${BOLD}RikkaHub → Kelivo${NC}    将 RikkaHub 备份转换为 Kelivo 格式"
    echo -e "    ${GREEN}3)${NC}  ${BOLD}查看备份信息${NC}         分析备份文件内容（不转换）"
    echo -e "    ${GREEN}4)${NC}  ${BOLD}自动检测并转换${NC}       自动识别备份格式并双向转换"
    echo ""
    echo -e "    ${GREEN}0)${NC}  ${BOLD}退出${NC}"
    echo ""
}

# 选择文件（交互式）
choose_file() {
    local prompt="$1"
    local file=""

    echo -e "  ${prompt}"
    echo -e "  ${YELLOW}请输入文件路径（支持拖拽文件到终端）：${NC}"
    echo -n "  > "
    read -r file

    # 去掉可能的引号（拖拽文件时可能有）
    file="${file#\'}"
    file="${file%\'\"}"
    file="${file#\"}"
    file="${file%\"}"

    if [ -z "$file" ]; then
        echo -e "  ${RED}错误：未输入文件路径${NC}"
        return 1
    fi

    if [ ! -f "$file" ]; then
        echo -e "  ${RED}错误：文件不存在 - $file${NC}"
        return 1
    fi

    echo "$file"
}

# 选择输出文件（可选）
choose_output() {
    local file=""

    echo -e "  ${YELLOW}请输入输出文件路径（按 Enter 使用默认路径）：${NC}"
    echo -n "  > "
    read -r file

    file="${file#\'}"
    file="${file%\'\"}"
    file="${file#\"}"
    file="${file%\"}"

    echo "$file"
}

# 从当前目录查找备份文件
find_backups() {
    local dir="${1:-.}"
    local files=()
    local i=1

    echo ""
    echo -e "  ${CYAN}当前目录下的备份文件：${NC}"
    echo ""

    while IFS= read -r -d '' f; do
        files+=("$f")
        local name=$(basename "$f")
        local size=$(ls -lh "$f" 2>/dev/null | awk '{print $5}')
        echo -e "    ${GREEN}${i})${NC} ${name}  ${YELLOW}(${size})${NC}"
        i=$((i + 1))
    done < <(find "$dir" -maxdepth 1 -name "*.zip" -print0 2>/dev/null)

    if [ ${#files[@]} -eq 0 ]; then
        echo -e "    ${YELLOW}未找到 .zip 文件${NC}"
        return 1
    fi

    echo ""
    echo -e "  ${YELLOW}请选择文件编号（或输入 0 手动输入路径）：${NC}"
    echo -n "  > "
    read -r choice

    if [ "$choice" = "0" ] || [ -z "$choice" ]; then
        return 1
    fi

    if [ "$choice" -ge 1 ] 2>/dev/null && [ "$choice" -le ${#files[@]} ] 2>/dev/null; then
        echo "${files[$((choice - 1))]}"
    else
        echo -e "  ${RED}无效选择${NC}"
        return 1
    fi
}

# 执行转换
do_convert() {
    local mode="$1"
    local input=""
    local output=""

    # 先尝试列出目录中的文件
    input=$(find_backups "$SCRIPT_DIR") || true
    if [ -z "$input" ]; then
        input=$(choose_file "请选择输入备份文件") || return 1
    fi

    echo ""
    local input_name=$(basename "$input")
    echo -e "  ${GREEN}已选择文件：${input_name}${NC}"
    echo ""

    # 询问输出路径
    echo -e "  ${YELLOW}指定输出文件路径？（按 Enter 自动生成）${NC}"
    echo -n "  > "
    read -r output
    output="${output#\'}"
    output="${output%\'\"}"
    output="${output#\"}"
    output="${output%\"}"

    echo ""
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    if [ -n "$output" ]; then
        "$CONVERTER" "$mode" -i "$input" -o "$output"
    else
        "$CONVERTER" "$mode" -i "$input"
    fi

    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 查看备份信息
do_info() {
    local input=""

    input=$(find_backups "$SCRIPT_DIR") || true
    if [ -z "$input" ]; then
        input=$(choose_file "请选择要查看的备份文件") || return 1
    fi

    echo ""
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    "$CONVERTER" info -i "$input"
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 自动检测并转换
do_auto_convert() {
    local input=""
    local output=""

    input=$(find_backups "$SCRIPT_DIR") || true
    if [ -z "$input" ]; then
        input=$(choose_file "请选择要转换的备份文件") || return 1
    fi

    # 检测格式
    local format=""
    if unzip -l "$input" 2>/dev/null | grep -q "rikka_hub.db"; then
        format="rikkahub"
    elif unzip -l "$input" 2>/dev/null | grep -q "chats.json"; then
        format="kelivo"
    fi

    echo ""
    local input_name=$(basename "$input")
    echo -e "  ${GREEN}已选择文件：${input_name}${NC}"
    echo -e "  ${GREEN}检测格式：${format}${NC}"
    echo ""

    # 询问输出路径
    echo -e "  ${YELLOW}指定输出文件路径？（按 Enter 自动生成）${NC}"
    echo -n "  > "
    read -r output
    output="${output#\'}"
    output="${output%\'\"}"
    output="${output#\"}"
    output="${output%\"}"

    echo ""
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    case "$format" in
        "rikkahub")
            if [ -n "$output" ]; then
                "$CONVERTER" rikkahub2kelivo -i "$input" -o "$output"
            else
                "$CONVERTER" rikkahub2kelivo -i "$input"
            fi
            ;;
        "kelivo")
            if [ -n "$output" ]; then
                "$CONVERTER" kelivo2rikkahub -i "$input" -o "$output"
            else
                "$CONVERTER" kelivo2rikkahub -i "$input"
            fi
            ;;
        *)
            echo -e "  ${RED}无法识别备份格式，请手动选择转换方向${NC}"
            ;;
    esac

    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 主循环
main() {
    check_binary

    while true; do
        show_title
        show_menu

        echo -n "  请输入选项 [0-4]: "
        read -r choice
        echo ""

        case $choice in
            1)
                do_convert "kelivo2rikkahub"
                echo ""
                echo -e "  ${YELLOW}按 Enter 返回主菜单...${NC}"
                read -r
                ;;
            2)
                do_convert "rikkahub2kelivo"
                echo ""
                echo -e "  ${YELLOW}按 Enter 返回主菜单...${NC}"
                read -r
                ;;
            3)
                do_info
                echo ""
                echo -e "  ${YELLOW}按 Enter 返回主菜单...${NC}"
                read -r
                ;;
            4)
                do_auto_convert
                echo ""
                echo -e "  ${YELLOW}按 Enter 返回主菜单...${NC}"
                read -r
                ;;
            0)
                echo -e "  ${GREEN}再见！${NC}"
                echo ""
                exit 0
                ;;
            *)
                echo -e "  ${RED}无效选项，请重新选择${NC}"
                sleep 1
                ;;
        esac
    done
}

main
