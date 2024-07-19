FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /app/main .

USER 65532:65532

ENTRYPOINT ["./main"]