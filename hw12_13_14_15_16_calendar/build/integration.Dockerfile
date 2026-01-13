# Собираем в гошке
FROM golang:1.24 as build

WORKDIR /app

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go test -c -o integration-tests ./internal/integration/...

# На выходе тонкий образ
FROM alpine:3.9 as final

LABEL SERVICE="integration"

WORKDIR /app

COPY --from=build /app/integration-tests ./integration-tests

RUN chmod +x ./integration-tests

ENTRYPOINT ["./integration-tests"]
CMD ["-test.v"]