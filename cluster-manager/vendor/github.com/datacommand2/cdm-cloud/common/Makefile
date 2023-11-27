.PHONY: vendor test lint clean

vendor:
	GOFLAGS='-mod=vendor' go mod tidy
	GOFLAGS='-mod=vendor' go mod vendor

test:
	GOFLAGS='-mod=vendor' go test -p 1 -v -race -coverprofile=coverage.out $$(go list ./... | grep -v /vendor/)
	GOFLAGS='-mod=vendor' go tool cover -func=coverage.out
	GOFLAGS='-mod=vendor' go tool cover -html=coverage.out -o coverage.html

lint:
	go fmt $$(go list ./... | grep -v /vendor/)
	go vet $$(go list ./... | grep -v /vendor/)
	GOFLAGS='-mod=vendor' golint -set_exit_status $$(go list ./... | grep -v /vendor/) || exit 1

clean:
	@rm -f coverage.out coverage.html
