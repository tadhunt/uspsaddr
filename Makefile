all: generate tidy fmt vet test

generate:
	mkdir -p uspsinternal
	oapi-codegen -config oapi-codegen.yaml usps-addresses-v3r2_2.yaml

clean:
	rm -rf uspsinternal

fmt:
	go fmt ./...

vet:
	go vet ./...
	staticcheck


test:
	go test -v ./...

tidy:
	go mod tidy
