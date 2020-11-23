APP=identification-service
APP_VERSION:=0.1
APP_COMMIT:=$(shell git rev-parse HEAD)
APP_EXECUTABLE="./out/$(APP)"
ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")

LOCAL_CONFIG_FILE=local.env
DOCKER_REGISTRY_USER_NAME=nsnikhil
HTTP_SERVE_COMMAND=http-serve
WORKER_COMMAND=worker
MIGRATE_COMMAND=migrate
ROLLBACK_COMMAND=rollback

setup: copy-config init-db init-test-db migrate test

init-db:
	psql -c "create user id_user superuser password 'id_password';" -U postgres
	psql -c "create database id_db owner=id_user" -U postgres

init-test-db:
	psql -c "create user id_test_user superuser password 'id_test_password';" -U postgres
	psql -c "create database id_test_db owner=id_test_user" -U postgres

deps:
	go mod download

tidy:
	go mod tidy

check: fmt vet lint

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

lint:
	golint $(ALL_PACKAGES)

compile:
	mkdir -p out/
	go build -ldflags "-X main.version=$(APP_VERSION) -X main.commit=$(APP_COMMIT)" -o $(APP_EXECUTABLE) cmd/*.go

build: deps compile

local-http-serve: build
	$(APP_EXECUTABLE) -configFile=$(LOCAL_CONFIG_FILE) $(HTTP_SERVE_COMMAND)

http-serve: build
	$(APP_EXECUTABLE) -configFile=$(configFile) $(HTTP_SERVE_COMMAND)

local-worker: build
	$(APP_EXECUTABLE) -configFile=$(LOCAL_CONFIG_FILE) $(WORKER_COMMAND)

worker: build
	$(APP_EXECUTABLE) -configFile=$(configFile) $(WORKER_COMMAND)

docker-build:
	docker build -t $(DOCKER_REGISTRY_USER_NAME)/$(APP):$(APP_VERSION) .
	docker rmi -f $$(docker images -f "dangling=true" -q)

docker-push: docker-build
	docker push $(DOCKER_REGISTRY_USER_NAME)/$(APP):$(APP_VERSION)

clean:
	rm -rf out/

copy-config:
	cp .env.sample local.env

test:
	go clean -testcache
	go test -p 1 ./...

ci-test: copy-config init-db migrate test

test-cover-html:
	go clean -testcache
	mkdir -p out/
	go test -p 1 ./... -coverprofile=out/coverage.out
	go tool cover -html=out/coverage.out

migrate: build
	$(APP_EXECUTABLE) $(MIGRATE_COMMAND)

rollback: build
	$(APP_EXECUTABLE) $(ROLLBACK_COMMAND)