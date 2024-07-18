FROM golang:1.21

RUN mkdir /app
WORKDIR /app

COPY . .
# COPY .env .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

EXPOSE 3200

CMD ["/main"]