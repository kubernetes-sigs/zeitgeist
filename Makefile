build:
	go build

lint:
	test -z $(shell go fmt .) || (echo "Linting failed !" && exit 8)
	go vet ./...
	GO111MODULE=off go get -u golang.org/x/lint/golint
	golint ./...

test:
	go test ./... -covermode=count -coverprofile=coverage.out

test-results: test
	go tool cover -html=coverage.out
