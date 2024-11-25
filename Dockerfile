FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN cd cmd/backup-creator && \
go build -o /app/backup-creator

FROM ubuntu:latest

RUN apt update -y && apt install -y ca-certificates

WORKDIR /

COPY --from=builder /app/backup-creator .

RUN chmod +x backup-creator

EXPOSE 8080

ENTRYPOINT ["./backup-creator"]
