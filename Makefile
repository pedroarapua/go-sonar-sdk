SHELL := /bin/bash
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_SHA = $(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_NUMBER = $(CIRCLE_BUILD_NUM)
VERSION = $(shell git describe --tags || echo "DEV")
GITHUB_USERNAME = luizalabs
COVERAGE_FILE = coverage.out 

setup:
	go get -u github.com/google/go-querystring/query

install:
	@dep ensure

nv: 
	@echo $(GOPACKAGES)

run:
	go run examples/simple/main.go

lint:
	golangci-lint run

test:
	go test `go list ./... | grep -v examples`

test-coverage:
	go test `go list ./... | grep -v examples` -race -coverprofile=$(COVERAGE_FILE)

coverage-html: test-coverage
	go tool cover -html=${COVERAGE_FILE}

coverage-missing: gocov test-coverage
	gocov convert ${COVERAGE_FILE} | gocov annotate - | grep MISS

send-sonar-ci:
	sonar-scanner -Dsonar.projectVersion=$(VERSION)

send-sonar-local:
	sonar-scanner -Dsonar.projectVersion=$(VERSION) -Dsonar.login=$(CI_SONAR_TOKEN)

send-codecov-ci:
	bash <(curl -s https://codecov.io/bash)

send-codecov-local:
	bash <(curl -s https://codecov.io/bash) -t ${CI_CODECOV_TOKEN}

next-version: git-semver
	git semver next

release: next-version
	git push --tags

git-semver:
	git semver 1>/dev/null 2>&1 || (git clone https://github.com/markchalloner/git-semver.git /tmp/git-semver && cd /tmp/git-semver && git checkout $( \
    git tag | grep '^[0-9]\+\.[0-9]\+\.[0-9]\+$' | \
    sort -t. -k 1,1n -k 2,2n -k 3,3n | tail -n 1 \
) && sudo ./install.sh)