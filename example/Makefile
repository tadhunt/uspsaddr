all:
	go mod tidy
	go vet
	staticcheck
	go build -o example

test: export USPS_CLIENT_ID := $(shell jq -r .ClientID < ../secrets/config.json)
test: export USPS_CLIENT_SECRET := $(shell jq -r .ClientSecret < ../secrets/config.json)
test: all
	./example
