# Makefile to provide common wrapped commands

# go dep version https://github.com/golang/dep
DEP_VERSION=0.3.2

# check the undelaying operation system and set as lowercase for further processing
OS := $(shell uname | tr A-Z a-z)

# check if dep is currently installed
DEP := $(shell command -v dep 2> /dev/null)

# check if go dep (https://github.com/golang/dep) is installed; if not install and configure it
ensure_go_dep:
ifndef DEP
	@echo "go dep not installed; installing..."
	- curl -L -s https://github.com/golang/dep/releases/download/v$(DEP_VERSION)/dep-$(OS)-amd64 -o $GOPATH/bin/dep
	- chmod +x $(GOPATH)/bin/dep
endif
	@echo "go dep already installed."

ensure_coverall:
	- go get github.com/mattn/goveralls
	- go get github.com/modocache/gover

tests_with_cover:
	- go list -f '{{if len .TestGoFiles}}"go test -tags test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

goveralls:
	- gover
	- "$HOME/gopath/bin/goveralls -coverprofile=gover.coverprofile -service=travis-ci -repotoken $COVERALLS_TOKEN"