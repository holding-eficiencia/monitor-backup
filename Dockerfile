# Use a imagem oficial do Go como base
FROM golang:1.22.3 as builder

# Defina o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copie os arquivos go.mod e go.sum para o diretório de trabalho
COPY go.mod go.sum ./

# Baixe as dependências
RUN go mod download

# Copie o código-fonte do projeto para o diretório de trabalho
COPY . .

# Compile o aplicativo
RUN CGO_ENABLED=0 GOOS=linux go build -o backups-exporter

# Use uma imagem mais leve para rodar o aplicativo
FROM alpine:latest

# Instale as dependências necessárias
RUN apk --no-cache add ca-certificates

# Defina o diretório de trabalho
WORKDIR /root/

# Copie o binário compilado da imagem anterior
COPY --from=builder /app/backups-exporter .

# Exponha a porta em que o aplicativo será executado
EXPOSE 8085

# Comando para executar o aplicativo
CMD ["./backups-exporter"]
