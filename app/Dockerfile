FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./main.go

FROM alpine:3.22
WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /src/docs /app/docs

EXPOSE 8080
CMD ["/app/api"]
