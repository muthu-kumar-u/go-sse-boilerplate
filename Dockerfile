FROM golang:1.24-alpine AS builder

RUN apk add --no-cache tzdata
ENV TZ=Asia/Kolkata

WORKDIR /app/app-backend-v2
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o app-backend-v2 ./main.go

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Kolkata

WORKDIR /app
COPY --from=builder /app/app-backend-v2/app-backend-v2 ./

EXPOSE 8081
CMD ["./app-stream"]
