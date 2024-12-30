# 使用Go 1.22官方镜像作为基础镜像
FROM golang:1.22.3-bookworm

# 设置工作目录
WORKDIR /app

# 将Go模块文件和依赖项缓存到镜像中
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码到工作目录
COPY . .

# 编译Go应用程序
RUN go build -o main main.go

# 声明运行时端口
EXPOSE 8080

# 运行编译后的应用程序
CMD ["./main"]