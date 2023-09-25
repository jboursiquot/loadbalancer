default: 

tidy:
	go mod tidy

.PHONY: example
example:
	go run cmd/example/*.go
