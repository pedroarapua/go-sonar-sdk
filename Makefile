GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_SHA = $(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_NUMBER = $(CIRCLE_BUILD_NUM)
VERSION = $(shell git describe --tags || echo "DEV")
GITHUB_USERNAME = luizalabs
SONAR_TOKEN = $(CI_SONAR_TOKEN)

setup:
	go get -u github.com/google/go-querystring/query

install:
	@dep ensure

lint:
	golangci-lint run

test:
	go test `go list ./... | grep -v examples` -race -coverprofile=coverage.out -covermode=atomic

sonar:
	sonar-scanner -Dsonar.login=$(SONAR_TOKEN) -Dsonar.projectVersion=$(VERSION)

nv: 
	@echo $(GOPACKAGES)

clean:
	rm -drf dist/

run:
	go run examples/simple/main.go