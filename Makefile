PACKAGES=$(shell go list ./... | grep -v vendor/)

test:
	go test -v $(PACKAGES)

build:
	go build cmd/git-br/git-br.go

clean:
	rm -f git-br
