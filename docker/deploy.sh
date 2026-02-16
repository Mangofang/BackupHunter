#!/bin/bash

mkdir -p data/backend
[ ! -f data/backend/config.yaml ] && touch data/backend/config.yaml
chmod 755 data/backend/config.yaml

mkdir -p data/backend/{logs,files}
chmod 755 data/backend/{logs,files}

echo "开始构建并部署 BackupHunter..."
docker-compose up -d --build

echo "等待服务启动及配置文件生成 (约 10 秒)..."
sleep 10

CONFIG_FILE="./data/backend/config.yaml"

if [ -f "$CONFIG_FILE" ]; then
    echo ""
    echo "部署成功！初始账号密码如下："
    echo "========================================"
docker exec app_backend cat /app/config.yaml | grep -E "^\s+(username|password):"
    echo "========================================"
    echo ""
    echo "配置文件位置：$CONFIG_FILE"
    echo "访问地址：http://$(hostname -I | awk '{print $1}')"
    echo ""
    echo "安全提示：请妥善保管账号密码，建议首次登录后修改"
else
    echo "X 配置文件未生成，请检查后端日志："
    echo "   docker-compose logs backend"
fi
