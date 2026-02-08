default: build

build:
	go build -o terraform-provider-intune

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/MANCHTOOLS/tofutune/0.1.0/linux_amd64
	cp terraform-provider-intune ~/.terraform.d/plugins/registry.terraform.io/MANCHTOOLS/tofutune/0.1.0/linux_amd64/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

fmt:
	go fmt ./...
	terraform fmt -recursive

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download

generate:
	go generate ./...

clean:
	rm -f terraform-provider-intune

.PHONY: build install test testacc fmt lint deps generate clean
