# justfile

# Default recipe to run when just is called without arguments
default:
    @just --list

# Build and install hq
install: build
    sudo mv build/hq /usr/local/bin/

# Just build the binary
build:
    mkdir -p build
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -buildvcs=false -o build/hq

# Clean built binary
clean:
    rm -f hq

# Run tests
test:
    go test ./...

# Build and run locally without installing
run URL SELECTOR *ARGS:
    curl -s {{ URL }} | go run . {{ SELECTOR }} {{ ARGS }}

help:
    go run . --help

# Remove the installed binary
uninstall:
    sudo rm -f /usr/local/bin/hq

completion SHELL:
    go run . completion {{ SHELL }}