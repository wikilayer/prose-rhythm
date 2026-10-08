.DEFAULT_GOAL := build

STATICCHECK_VERSION ?= v0.8.1
LEADING := ../prose-rhythm-python

.PHONY: install-tools format comments lint test-build test build data measure

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	python3 -m pip install --quiet --upgrade git+https://github.com/botforge-pro/commentcensor.git

format:
	gofmt -w .

comments:
	commentcensor *.go

lint: comments
	go vet ./...
	gofmt -l . | (! grep .)
	staticcheck ./...

test-build:
	go build ./...
	go test -run '^$$' ./...

test:
	go test ./...

build: lint test-build test
	go build ./...

data:
	$(MAKE) -C $(LEADING) data
	mkdir -p data
	for id in $$(awk '/^  - id:/{print $$3}' corpus/novels.yaml); do \
		python3 $(LEADING)/tools/prose.py $(LEADING)/data/$$id.txt > data/$$id.txt || exit 1; \
	done

measure: data
	go test -tags novels -run Novel ./...
