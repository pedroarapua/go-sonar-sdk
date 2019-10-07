GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_SHA = $(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_NUMBER = $(CIRCLE_BUILD_NUM)
VERSION = $(shell git describe --tags || echo "DEV")
GITHUB_USERNAME = luizalabs

setup:
	go get -u github.com/google/go-querystring/query

install:
	@dep ensure

nv: 
	@echo $(GOPACKAGES)

clean:
	rm -drf dist/