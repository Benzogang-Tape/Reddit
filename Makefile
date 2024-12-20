APP_NAME = redditclone
APP_DEMO = demo

.PHONY:
.SILENT:
.DEFAULT_GOAL := run

.PHONY: build
build:
	go build -v -o ./bin/${APP_NAME} ./cmd/${APP_NAME}

.PHONY: build-demo
build-demo:
	go build -v -o ./bin/${APP_DEMO} ./cmd/${APP_DEMO}

.PHONY: run
run: swag
	docker-compose --env-file .env up -d
	#docker-compose --env APP_NAME=${APP_NAME} up -d

.PHONY: run-demo
run-demo: swag-demo build-demo
	./bin/${APP_DEMO}

.PHONY: lint
lint:
	golangci-lint run \
	-v -c .golangci.yaml --color='always' \
	--exclude-dirs-use-default --exclude-files './internal/storage/mocks/*','\*.mod','\*.sum' \
	--exclude-dirs 'vendor'

.PHONY: test
test: gen
	go test -v -coverprofile=./coverage/cover.out ./...
	make test.coverage

.PHONY: test.coverage
test.coverage:
	go tool cover -html=./coverage/cover.out -o ./coverage/cover.html
	go tool cover -func=./coverage/cover.out | grep "total"

.PHONY: gen
gen:
	go generate ./...

.PHONY: clean
clean:
	go clean
	rm -f ./bin/${APP_NAME} ./bin/${APP_DEMO}

.PHONY: swag
swag:
	swag fmt
	swag init -g ./cmd/${APP_NAME}/${APP_NAME}.go

.PHONY: swag-demo
swag-demo:
	swag fmt
	swag init -g ./cmd/${APP_DEMO}/${APP_DEMO}.go

#go test -v -cover ./... | grep 'ok' | sed 's/.* \([0-9.]\+\)%/\1/' | grep -oE '[0-9.]+' > coverage.txt

#go test -v -coverprofile=cover.out ./...
#go tool cover -func=cover.out -o cover.txt
#cat cover.txt | grep 'total:' | awk -F' ' '{print $$(NF)}' | sed 's/%//' > cover.txt
