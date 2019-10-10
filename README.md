# go-sonar-sdk
[![Build Status](https://travis-ci.com/pedroarapua/go-sonar-sdk.svg?token=wtkD8x3vzz1kYkvou9fn&branch=master)](https://travis-ci.com/luizalabs/techlead-metrics)
[![CircleCI](https://circleci.com/pedroarapua/go-sonar-sdk/tree/master.svg?style=svg&circle-token=440554fe43158ca84024dc4bd2a27e069fed0bc4)](https://circleci.com/pedroarapua/go-sonar-sdk/tree/master)
[![Coverage Status](https://coveralls.io/repos/github/pedroarapua/go-sonar-sdk/badge.svg?t=pWurmF)](https://coveralls.io/github/pedroarapua/go-sonar-sdk)
[![codecov](https://codecov.io/gh/pedroarapua/go-sonar-sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/pedroarapua/go-sonar-sdk)

go-sonar-sdk is a sdk to use sonar API.

## Usage

go-sonar-sdk requires or GoLang 1.9 later and glide.

To install tools
```
make setup
```

### Install Dependencies

To config dependencies in private repositories
```
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

```
make install
```

### Build

```
go build .
```
