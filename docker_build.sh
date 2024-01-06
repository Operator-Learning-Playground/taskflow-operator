#!/bin/bash

# 设置变量
BUILD_CONTEXT="./container-agent"  # 构建上下文目录

# 切换到构建上下文目录
cd "${BUILD_CONTEXT}" || exit
DOCKERFILE_PATH="Dockerfile"  # Dockerfile 的路径
IMAGE_NAME="docker.io/taskflow/agent" # 镜像名称
TAG="v1.0"                                      # 镜像标签


echo "镜像 ${IMAGE_NAME}:${TAG} 开始构建..."
# 执行 Docker build 命令
docker build -t "${IMAGE_NAME}:${TAG}" -f "${DOCKERFILE_PATH}" .
echo "镜像构建完成。"