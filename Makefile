SOURCES=$(shell find . -name "*.go" | grep -v vendor/)
PACKAGES=$(shell go list ./... | grep -v vendor/)
FGT := fgt

test:
	go test -v $(PACKAGES)

build:
	go build cmd/git-br/git-br.go

lint:
	$(FGT) go fmt $(PACKAGES)
	$(FGT) goimports -w $(SOURCES)
	go list ./... | grep -v vendor/ | xargs -L1 $(FGT) golint
	$(FGT) go vet $(PACKAGES)
	$(FGT) errcheck -ignore Close $(PACKAGES)
	staticcheck $(PACKAGES)
	$(FGT) gosimple $(PACKAGES)

lint-deps:
	go get -u github.com/GeertJohan/fgt
	go get -u golang.org/x/tools/cmd/cover
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u honnef.co/go/simple/cmd/gosimple
	go get -u github.com/mvdan/interfacer/cmd/interfacer
	go get -u honnef.co/go/tools/cmd/staticcheck

clean:
	rm -f git-br
