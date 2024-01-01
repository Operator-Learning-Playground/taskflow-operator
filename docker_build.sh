#!/bin/bash

# 设置变量
BUILD_CONTEXT="./container-agent"  # 构建上下文目录

# 切换到构建上下文目录
cd "${BUILD_CONTEXT}" || exit
DOCKERFILE_PATH="Dockerfile"  # Dockerfile 的路径
IMAGE_NAME="docker.io/taskflow/agent" # 镜像名称
TAG="v1.0"                                      # 镜像标签


# 检查镜像是否已经存在
if docker inspect "${IMAGE_NAME}:${TAG}" &> /dev/null; then
  echo "镜像 ${IMAGE_NAME}:${TAG} 已存在，无需构建。"
else
  echo "镜像 ${IMAGE_NAME}:${TAG} 不存在，开始构建..."
  # 执行 Docker build 命令
  docker build -t "${IMAGE_NAME}:${TAG}" -f "${DOCKERFILE_PATH}" .
  echo "镜像构建完成。"

  # 可选：推送镜像到 Docker 镜像仓库
  # docker push "${IMAGE_NAME}:${TAG}"
fi