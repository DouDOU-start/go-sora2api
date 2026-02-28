#!/bin/bash
#
# Sora2API 安装 & 管理脚本
# 安装: curl -sSL https://raw.githubusercontent.com/DouDOU-start/go-sora2api/master/deploy/install.sh | sudo bash
# 安装后直接使用: sudo sora2api status / upgrade / logs ...
#

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置
GITHUB_REPO="DouDOU-start/go-sora2api"
APP_NAME="sora2api"
INSTALL_DIR="/opt/sora2api"
CONFIG_DIR="/etc/sora2api"
SERVICE_USER="sora2api"
CLI_LINK="/usr/local/bin/sora2api"
DEFAULT_PORT=8686

# 运行时变量
SERVER_HOST="0.0.0.0"
SERVER_PORT="$DEFAULT_PORT"

# ============================================================
# 工具函数
# ============================================================

print_info()    { echo -e "${BLUE}[信息]${NC} $1"; }
print_success() { echo -e "${GREEN}[成功]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[警告]${NC} $1"; }
print_error()   { echo -e "${RED}[错误]${NC} $1"; }

is_interactive() {
    [ -e /dev/tty ] && [ -r /dev/tty ] && [ -w /dev/tty ]
}

validate_port() {
    local port="$1"
    [[ "$port" =~ ^[0-9]+$ ]] && [ "$port" -ge 1 ] && [ "$port" -le 65535 ]
}

# ============================================================
# 检查与检测
# ============================================================

check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        print_error "请使用 root 权限运行（sudo）"
        exit 1
    fi
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)       ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) print_error "不支持的架构: $ARCH"; exit 1 ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        *) print_error "不支持的操作系统: $OS"; exit 1 ;;
    esac

    print_info "检测到平台: ${OS}_${ARCH}"
}

check_dependencies() {
    local missing=()
    command -v curl &>/dev/null || missing+=("curl")
    command -v tar  &>/dev/null || missing+=("tar")

    if [ ${#missing[@]} -gt 0 ]; then
        print_error "缺少依赖: ${missing[*]}"
        print_info "请先安装以上依赖"
        exit 1
    fi
}

# ============================================================
# 版本管理
# ============================================================

get_latest_version() {
    print_info "正在获取最新版本..."
    LATEST_VERSION=$(curl -s --connect-timeout 10 --max-time 30 \
        "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_error "获取最新版本失败，请检查网络连接"
        exit 1
    fi
    print_info "最新版本: $LATEST_VERSION"
}

get_current_version() {
    if [ -f "$INSTALL_DIR/sora2api-server" ]; then
        "$INSTALL_DIR/sora2api-server" --version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown"
    else
        echo "not_installed"
    fi
}

validate_version() {
    local version="$1"
    if [ -z "$version" ]; then
        print_error "请指定版本号（例如: v1.2.0）"
        exit 1
    fi
    [[ "$version" =~ ^v ]] || version="v$version"

    print_info "正在验证版本 $version..." >&2
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 --max-time 30 \
        "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${version}" 2>/dev/null)

    if [ "$http_code" != "200" ]; then
        print_error "版本不存在: $version" >&2
        list_versions >&2
        exit 1
    fi
    echo "$version"
}

list_versions() {
    print_info "正在获取可用版本..."
    local versions
    versions=$(curl -s --connect-timeout 10 --max-time 30 \
        "https://api.github.com/repos/${GITHUB_REPO}/releases" 2>/dev/null \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | head -20)

    if [ -z "$versions" ]; then
        print_error "获取版本列表失败"
        exit 1
    fi

    echo ""
    echo "可用版本:"
    echo "----------------------------------------"
    echo "$versions" | while read -r v; do echo "  $v"; done
    echo "----------------------------------------"
    echo ""
}

# ============================================================
# 安装核心
# ============================================================

download_and_extract() {
    local version_num=${LATEST_VERSION#v}
    local archive_name="${APP_NAME}_${version_num}_${OS}_${ARCH}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${archive_name}"
    local checksum_url="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/checksums.txt"

    print_info "正在下载 ${archive_name}..."

    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT

    if ! curl -sL "$download_url" -o "$TEMP_DIR/$archive_name"; then
        print_error "下载失败"
        exit 1
    fi

    # 校验
    print_info "正在校验文件..."
    if curl -sL "$checksum_url" -o "$TEMP_DIR/checksums.txt" 2>/dev/null; then
        local expected=$(grep "$archive_name" "$TEMP_DIR/checksums.txt" | awk '{print $1}')
        local actual=$(sha256sum "$TEMP_DIR/$archive_name" | awk '{print $1}')
        if [ "$expected" != "$actual" ]; then
            print_error "校验失败"
            exit 1
        fi
        print_success "校验通过"
    else
        print_warning "无法验证校验和（checksums.txt 未找到）"
    fi

    # 解压
    print_info "正在解压..."
    tar -xzf "$TEMP_DIR/$archive_name" -C "$TEMP_DIR"

    mkdir -p "$INSTALL_DIR"
    cp "$TEMP_DIR/sora2api-server" "$INSTALL_DIR/sora2api-server"
    chmod +x "$INSTALL_DIR/sora2api-server"

    print_success "二进制文件已安装到 $INSTALL_DIR/sora2api-server"
}

create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        print_info "用户已存在: $SERVICE_USER"
    else
        print_info "正在创建系统用户 $SERVICE_USER..."
        useradd -r -s /bin/sh -d "$INSTALL_DIR" "$SERVICE_USER"
        print_success "用户已创建"
    fi
}

setup_directories() {
    print_info "正在设置目录..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"

    # 如果配置文件不存在，创建默认配置
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        cat > "$CONFIG_DIR/config.yaml" << YAML
server:
  host: "${SERVER_HOST}"
  port: ${SERVER_PORT}
  admin_user: "admin"
  admin_password: "admin123"

database:
  url: "postgres://postgres:postgres@localhost:5432/sora2api?sslmode=disable"
  log_level: "warn"
  auto_migrate: true
YAML
        print_info "已创建默认配置: $CONFIG_DIR/config.yaml"
    fi

    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR"
    print_success "目录配置完成"
}

install_service() {
    print_info "正在安装 systemd 服务..."

    cat > /etc/systemd/system/sora2api.service << EOF
[Unit]
Description=Sora2API - Sora API Gateway
Documentation=https://github.com/DouDOU-start/go-sora2api
After=network.target postgresql.service

[Service]
Type=simple
User=sora2api
Group=sora2api
WorkingDirectory=/opt/sora2api
ExecStart=/opt/sora2api/sora2api-server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sora2api

NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/opt/sora2api /etc/sora2api

Environment=GIN_MODE=release
Environment=CONFIG_PATH=/etc/sora2api/config.yaml

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    print_success "systemd 服务已安装"
}

# 安装管理命令到 /usr/local/bin/sora2api
install_cli() {
    print_info "正在安装管理命令..."
    # 复制脚本自身到安装目录
    cp "$0" "$INSTALL_DIR/manage.sh" 2>/dev/null || {
        # 通过 pipe 执行时 $0 不可用，从远程下载
        curl -sL "https://raw.githubusercontent.com/${GITHUB_REPO}/master/deploy/install.sh" \
            -o "$INSTALL_DIR/manage.sh"
    }
    chmod +x "$INSTALL_DIR/manage.sh"

    # 创建 /usr/local/bin/sora2api 符号链接
    ln -sf "$INSTALL_DIR/manage.sh" "$CLI_LINK"
    print_success "管理命令已安装: sora2api"
}

configure_server() {
    if ! is_interactive; then
        print_info "服务器配置: ${SERVER_HOST}:${SERVER_PORT}（默认）"
        return
    fi

    echo ""
    echo -e "${CYAN}=============================================="
    echo "  服务器配置"
    echo "==============================================${NC}"
    echo ""

    echo -e "${YELLOW}0.0.0.0 表示监听所有网卡，127.0.0.1 仅本地访问${NC}"
    read -p "监听地址 [${SERVER_HOST}]: " input_host < /dev/tty
    [ -n "$input_host" ] && SERVER_HOST="$input_host"

    echo ""
    echo -e "${YELLOW}建议使用 1024-65535 之间的端口${NC}"
    while true; do
        read -p "端口 [${SERVER_PORT}]: " input_port < /dev/tty
        if [ -z "$input_port" ]; then
            break
        elif validate_port "$input_port"; then
            SERVER_PORT="$input_port"
            break
        else
            print_error "无效端口号，请输入 1-65535 之间的数字"
        fi
    done

    echo ""
    print_info "服务器配置: ${SERVER_HOST}:${SERVER_PORT}"
}

get_public_ip() {
    print_info "正在获取公网 IP..."
    local response
    response=$(curl -s --connect-timeout 5 --max-time 10 "https://ipinfo.io/json" 2>/dev/null)
    if [ -n "$response" ]; then
        PUBLIC_IP=$(echo "$response" | grep -o '"ip": *"[^"]*"' | sed 's/"ip": *"\([^"]*\)"/\1/')
        if [ -n "$PUBLIC_IP" ]; then
            return 0
        fi
    fi
    print_warning "无法获取公网 IP，使用本地 IP"
    PUBLIC_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "YOUR_SERVER_IP")
}

start_and_enable() {
    print_info "正在启动服务..."
    if systemctl start sora2api; then
        print_success "服务已启动"
    else
        print_error "服务启动失败，请检查日志: sudo sora2api logs"
    fi

    systemctl enable sora2api 2>/dev/null && print_success "已设置开机自启"
}

# ============================================================
# 服务管理快捷命令
# ============================================================

cmd_status() {
    echo ""
    echo -e "${CYAN}Sora2API 服务状态${NC}"
    echo "=============================================="
    echo ""

    # 版本信息
    local ver=$(get_current_version)
    echo -e "  版本:    ${GREEN}${ver}${NC}"
    echo -e "  安装目录: $INSTALL_DIR"
    echo -e "  配置文件: $CONFIG_DIR/config.yaml"
    echo ""

    # systemd 状态
    systemctl status sora2api --no-pager 2>/dev/null || print_warning "服务未安装"
}

cmd_logs() {
    local lines="${1:-50}"
    journalctl -u sora2api -n "$lines" --no-pager -o short-iso
}

cmd_logs_follow() {
    journalctl -u sora2api -f -o short-iso
}

cmd_start() {
    check_root
    print_info "正在启动服务..."
    if systemctl start sora2api; then
        print_success "服务已启动"
    else
        print_error "启动失败，请查看日志: sudo sora2api logs"
    fi
}

cmd_stop() {
    check_root
    print_info "正在停止服务..."
    systemctl stop sora2api
    print_success "服务已停止"
}

cmd_restart() {
    check_root
    print_info "正在重启服务..."
    if systemctl restart sora2api; then
        print_success "服务已重启"
    else
        print_error "重启失败，请查看日志: sudo sora2api logs"
    fi
}

cmd_config() {
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        ${EDITOR:-vi} "$CONFIG_DIR/config.yaml"
    else
        print_error "配置文件不存在: $CONFIG_DIR/config.yaml"
    fi
}

cmd_version() {
    local ver=$(get_current_version)
    echo "sora2api $ver"
}

# ============================================================
# 安装完成提示
# ============================================================

print_completion() {
    local display_host="${PUBLIC_IP:-YOUR_SERVER_IP}"
    [ "$SERVER_HOST" = "127.0.0.1" ] && display_host="127.0.0.1"

    echo ""
    echo "=============================================="
    print_success "Sora2API 安装完成！"
    echo "=============================================="
    echo ""
    echo "  安装目录: $INSTALL_DIR"
    echo "  配置文件: $CONFIG_DIR/config.yaml"
    echo "  监听地址: ${SERVER_HOST}:${SERVER_PORT}"
    echo ""
    echo "  访问地址: http://${display_host}:${SERVER_PORT}"
    echo ""
    echo "=============================================="
    echo "  管理命令（sudo sora2api <命令>）"
    echo "=============================================="
    echo ""
    echo "  sora2api status          查看服务状态"
    echo "  sora2api start           启动服务"
    echo "  sora2api stop            停止服务"
    echo "  sora2api restart         重启服务"
    echo "  sora2api logs            查看最近日志"
    echo "  sora2api logs -f         实时跟踪日志"
    echo "  sora2api config          编辑配置文件"
    echo "  sora2api version         查看当前版本"
    echo "  sora2api upgrade         升级到最新版本"
    echo "  sora2api upgrade -v x.x  升级到指定版本"
    echo "  sora2api list-versions   列出可用版本"
    echo "  sora2api uninstall       卸载"
    echo ""
    echo "=============================================="
}

# ============================================================
# 升级
# ============================================================

upgrade() {
    if [ ! -f "$INSTALL_DIR/sora2api-server" ]; then
        print_error "Sora2API 尚未安装，请先执行全新安装"
        exit 1
    fi

    print_info "正在升级 Sora2API..."
    local current=$(get_current_version)
    print_info "当前版本: $current"

    systemctl is-active --quiet sora2api && {
        print_info "正在停止服务..."
        systemctl stop sora2api
    }

    cp "$INSTALL_DIR/sora2api-server" "$INSTALL_DIR/sora2api-server.backup"
    print_info "备份已创建: $INSTALL_DIR/sora2api-server.backup"

    get_latest_version
    download_and_extract
    chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/sora2api-server"

    # 同时更新管理脚本
    install_cli

    print_info "正在启动服务..."
    systemctl start sora2api

    local new_ver=$(get_current_version)
    echo ""
    print_success "升级完成！ $current -> $new_ver"
}

# ============================================================
# 安装指定版本
# ============================================================

install_version() {
    local target_version="$1"

    if [ ! -f "$INSTALL_DIR/sora2api-server" ]; then
        print_error "Sora2API 尚未安装，请先执行全新安装"
        exit 1
    fi

    target_version=$(validate_version "$target_version")
    print_info "正在安装指定版本: $target_version"

    local current=$(get_current_version)
    print_info "当前版本: $current"

    if [ "$current" = "$target_version" ] || [ "$current" = "${target_version#v}" ]; then
        print_warning "已经是该版本，无需操作"
        exit 0
    fi

    systemctl is-active --quiet sora2api && {
        print_info "正在停止服务..."
        systemctl stop sora2api
    }

    cp "$INSTALL_DIR/sora2api-server" "$INSTALL_DIR/sora2api-server.backup"
    print_info "备份已创建"

    LATEST_VERSION="$target_version"
    download_and_extract
    chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/sora2api-server"

    print_info "正在启动服务..."
    systemctl start sora2api && print_success "服务已启动" || print_error "服务启动失败"

    echo ""
    print_success "指定版本安装完成！当前版本: $(get_current_version)"
}

# ============================================================
# 卸载
# ============================================================

uninstall() {
    print_warning "这将从系统中移除 Sora2API。"

    if ! is_interactive; then
        if [ "${FORCE_YES:-}" != "true" ]; then
            print_error "非交互模式，请使用 -y 确认"
            exit 1
        fi
    else
        read -p "确定要继续吗？(y/N) " -n 1 -r < /dev/tty
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "卸载已取消"
            exit 0
        fi
    fi

    systemctl stop sora2api 2>/dev/null || true
    systemctl disable sora2api 2>/dev/null || true

    rm -f /etc/systemd/system/sora2api.service
    systemctl daemon-reload

    # 移除管理命令
    rm -f "$CLI_LINK"

    rm -rf "$INSTALL_DIR"
    userdel "$SERVICE_USER" 2>/dev/null || true

    # 询问是否删除配置
    local remove_config=false
    if [ "${PURGE:-}" = "true" ]; then
        remove_config=true
    elif is_interactive; then
        read -p "是否同时删除配置目录 $CONFIG_DIR？[y/N]: " -n 1 -r < /dev/tty
        echo
        [[ $REPLY =~ ^[Yy]$ ]] && remove_config=true
    fi

    if [ "$remove_config" = true ]; then
        rm -rf "$CONFIG_DIR"
    else
        print_warning "配置目录未删除: $CONFIG_DIR"
    fi

    print_success "Sora2API 已卸载"
}

# ============================================================
# 主入口
# ============================================================

main() {
    local target_version=""
    local positional_args=()

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -y|--yes)    FORCE_YES="true"; shift ;;
            --purge)     PURGE="true"; shift ;;
            -v|--version)
                [ -n "${2:-}" ] && [[ ! "$2" =~ ^- ]] || { echo "错误: --version 需要版本参数"; exit 1; }
                target_version="$2"; shift 2 ;;
            --version=*) target_version="${1#*=}"; shift ;;
            *)           positional_args+=("$1"); shift ;;
        esac
    done
    set -- "${positional_args[@]}"

    case "${1:-}" in
        # ---- 服务管理命令（无需 banner）----
        status)
            cmd_status
            ;;
        start)
            cmd_start
            ;;
        stop)
            cmd_stop
            ;;
        restart)
            cmd_restart
            ;;
        logs)
            if [ "${2:-}" = "-f" ] || [ "${2:-}" = "--follow" ]; then
                cmd_logs_follow
            else
                cmd_logs "${2:-50}"
            fi
            ;;
        config)
            cmd_config
            ;;
        version)
            cmd_version
            ;;

        # ---- 安装/升级/卸载命令 ----
        upgrade|update)
            echo ""
            echo "=============================================="
            echo "       Sora2API 升级"
            echo "=============================================="
            echo ""
            check_root; detect_platform; check_dependencies
            if [ -n "$target_version" ]; then
                install_version "$target_version"
            else
                upgrade
            fi
            ;;
        install)
            echo ""
            echo "=============================================="
            echo "       Sora2API 安装"
            echo "=============================================="
            echo ""
            check_root; detect_platform; check_dependencies
            if [ -n "$target_version" ]; then
                if [ -f "$INSTALL_DIR/sora2api-server" ]; then
                    install_version "$target_version"
                else
                    configure_server
                    LATEST_VERSION=$(validate_version "$target_version")
                    download_and_extract; create_user; setup_directories; install_service; install_cli
                    get_public_ip; start_and_enable; print_completion
                fi
            else
                configure_server; get_latest_version
                download_and_extract; create_user; setup_directories; install_service; install_cli
                get_public_ip; start_and_enable; print_completion
            fi
            ;;
        list-versions|versions)
            list_versions
            ;;
        uninstall|remove)
            check_root; uninstall
            ;;
        --help|-h|help)
            echo ""
            echo "Sora2API 管理工具"
            echo ""
            echo "用法: sora2api <命令> [选项]"
            echo ""
            echo "服务管理:"
            echo "  status             查看服务状态和版本信息"
            echo "  start              启动服务"
            echo "  stop               停止服务"
            echo "  restart            重启服务"
            echo "  logs [N]           查看最近 N 条日志（默认 50）"
            echo "  logs -f            实时跟踪日志"
            echo "  config             编辑配置文件"
            echo "  version            显示当前版本"
            echo ""
            echo "安装与升级:"
            echo "  install            安装 Sora2API"
            echo "  upgrade            升级到最新版本"
            echo "  upgrade -v <ver>   升级到指定版本"
            echo "  list-versions      列出可用版本"
            echo "  uninstall          卸载 Sora2API"
            echo ""
            echo "选项:"
            echo "  -v, --version      指定版本号（例如: v1.2.0）"
            echo "  -y, --yes          跳过确认提示"
            echo ""
            ;;
        "")
            # 无参数：首次安装
            echo ""
            echo "=============================================="
            echo "       Sora2API 安装"
            echo "=============================================="
            echo ""
            check_root; detect_platform; check_dependencies
            if [ -n "$target_version" ]; then
                if [ -f "$INSTALL_DIR/sora2api-server" ]; then
                    install_version "$target_version"
                else
                    configure_server
                    LATEST_VERSION=$(validate_version "$target_version")
                    download_and_extract; create_user; setup_directories; install_service; install_cli
                    get_public_ip; start_and_enable; print_completion
                fi
            else
                configure_server; get_latest_version
                download_and_extract; create_user; setup_directories; install_service; install_cli
                get_public_ip; start_and_enable; print_completion
            fi
            ;;
        *)
            print_error "未知命令: $1"
            echo "运行 'sora2api help' 查看帮助"
            exit 1
            ;;
    esac
}

main "$@"
