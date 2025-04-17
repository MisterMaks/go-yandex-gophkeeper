FROM golang:1.23.4 AS dependencies
WORKDIR /go/src/go-yandex-gophkeeper
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS build
COPY . /go/src/go-yandex-gophkeeper
WORKDIR /go/src/go-yandex-gophkeeper
RUN go build -o ./bin/gophkeeper_server ./cmd/server

FROM debian
WORKDIR /app
COPY --from=build /go/src/go-yandex-gophkeeper/migrations/postgres/ /app/migrations/postgres
COPY --from=build /go/src/go-yandex-gophkeeper/certificates/ /app/certificates
COPY --from=build /go/src/go-yandex-gophkeeper/bin/gophkeeper_server /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/gophkeeper_server
