FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o /out/server ./cmd/server

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/server /app/server

EXPOSE 8080

USER app

CMD ["/app/server"]
