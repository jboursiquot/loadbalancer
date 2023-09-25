default: 

tidy:
	go mod tidy

.PHONY: example
example:
	go run cmd/example/*.go

.PHONY: lb
lb:
	go run cmd/lb/*.go

.PHONY: list-running-servers
list-running-servers:
	-@lsof -Pi :8080-8083