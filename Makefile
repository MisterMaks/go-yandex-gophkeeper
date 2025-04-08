.PHONY: build-server
build-server:
	@echo "-- building server"
	go build -o ./bin/go-yandex-gophkeeper-server ./cmd/server/

.PHONY: run-server
run-server:
	@echo "-- running server"
	./bin/go-yandex-gophkeeper-server

.PHONY: docker-server
docker-server:
	@echo "-- building docker container for server"
	docker build -f build/server.Dockerfile -t go-yandex-gophkeeper-server .

.PHONY: docker-run-server
docker-run-server:
	@echo "-- starting docker container for server"
	docker run -it -p 8080:8080 go-yandex-gophkeeper-server

.PHONY: dcb-server
dcb-server:
	@echo "-- starting docker compose for server"
	docker-compose -f ./deployments/server-docker-compose.yml up --build

.PHONY: build-clients
build-clients:
	@echo "-- building clients"
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.BuildVersion=v1.0.0 -X 'main.BuildDate=$$(date +'%Y/%m/%d')'" -o ./bin/go-yandex-gophkeeper-client-osx-arm ./cmd/client/
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.BuildVersion=v1.0.0 -X 'main.BuildDate=$$(date +'%Y/%m/%d')'" -o ./bin/go-yandex-gophkeeper-client-osx-x86 ./cmd/client/
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.BuildVersion=v1.0.0 -X 'main.BuildDate=$$(date +'%Y/%m/%d')'" -o ./bin/go-yandex-gophkeeper-client-windows.exe ./cmd/client/
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.BuildVersion=v1.0.0 -X 'main.BuildDate=$$(date +'%Y/%m/%d')'" -o ./bin/go-yandex-gophkeeper-client-linux ./cmd/client/

.PHONY: build-client
build-client:
	@echo "-- building clients"
	go build -o ./bin/go-yandex-gophkeeper-client ./cmd/client/

.PHONY: run-client
run-client:
	@echo "-- running client"
	./bin/go-yandex-gophkeeper-client

.PHONY: protoc
protoc:
	@echo "-- compiling .proto files"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
	--go-grpc_opt=paths=source_relative api/proto/service/service.proto;

.PHONY: godoc
godoc:
	@echo "-- running godoc server"
	godoc -http=:8000

.PHONY: test
test:
	@echo "-- testing \n-- env 'TEST_DATABASE_URI' required for integration tests"
	go test ./... -coverprofile cover.out

.PHONY: unit-test
unit-test:
	@echo "-- unit testing"
	go test ./... -short -coverprofile cover.out

.PHONY: test-cover
test-cover:
	@echo "-- testing with cover"
	go tool cover -html cover.out

.PHONY: test-cover-percentage
test-cover-percentage:
	@echo "-- testing with cover percentage"
	go test -v \
		-coverpkg=$(go list ./... | grep -v api/ | grep -v cmd/) \
		-coverprofile=cover.out \
		-covermode=count \
		$$(go list ./... | grep -v api/ | grep -v cmd/) && \
		go tool cover -func cover.out