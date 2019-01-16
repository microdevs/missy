
.PHONY: test lint format goveralls

lint:
	@if [ -n "$$(goimports -l */*.go)" ]; then \
		echo "Go code is not formatted:" ; \
		goimports -l *.go */*.go ; \
		exit 1; \
	fi

test:
	go test ./... --cover

test_goveralls:
	go list -f '{{if len .TestGoFiles}}"go test ./... -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

goveralls:
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	gover
	goveralls -coverprofile=gover.coverprofile -service=travis-ci -repotoken $(COVERALLS_TOKEN)

build:
	go build ./...
