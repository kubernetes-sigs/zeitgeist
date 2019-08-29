build:
	go build

test:
	go test ./... -covermode=count -coverprofile=coverage.out

test-results: test
	go tool cover -html=coverage.out
