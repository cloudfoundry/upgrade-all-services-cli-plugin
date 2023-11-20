###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help

help:  ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Test ###################################################################
.PHONY: test
test: download lint test-units ## run lint and unit tests

.PHONY: test-units
test-units:
	go run github.com/onsi/ginkgo/v2/ginkgo -r -p

.PHONY: download
download: ## download go module dependencies
	go mod download

###### Lint ###################################################################

.PHONY: lint
lint: checkformat checkimports vet staticcheck ## lint the source

checkformat: ## check that the code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then                   \
    		echo "gofmt check failed: run 'gofmt -d -e -l -w .'"; \
    		exit 1;                                               \
      fi

checkimports: ## check that imports are formatted correctly
	@@if [ -n "$$(go run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make format'";               \
		exit 1;                                                         \
	fi

vet: ## run go vet
	go vet ./...

staticcheck: ## run staticcheck
	go run honnef.co/go/tools/cmd/staticcheck ./...


###### Format #################################################################

.PHONY: format
format: ## format the source
	gofmt -s -e -l -w .
	go run golang.org/x/tools/cmd/goimports -l -w .

###### Build ##################################################################

.PHONY: build
build: ## use goreleaser to build the plugin for all target platforms
	goreleaser build --rm-dist --snapshot
