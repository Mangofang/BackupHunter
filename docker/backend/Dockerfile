FROM ubuntu:22.04

ENV TZ=Asia/Shanghai
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    tzdata \
    python3 \
    python3-pip \
    git \
    curl \
    vim \
    && rm -rf /var/lib/apt/lists/* \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

# 克隆 OneForAll 并安装依赖
RUN git clone https://github.com/shmilylty/OneForAll.git /app/OneForAll \
    && cd /app/OneForAll \
    && pip3 install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple --no-cache-dir

# 创建后端所需目录
RUN mkdir -p /app/data/logs /app/data/files

# 复制后端二进制文件 BackupHunter
COPY ./BackupHunter /app/BackupHunter
RUN chmod +x /app/BackupHunter

ENV DB_HOST=mysql
ENV DB_PORT=3306
ENV DB_USER=root
ENV DB_PASS=root
ENV DB_NAME=backup_hunter

# 启动命令使用 BackupHunter
CMD ["/app/BackupHunter"]
