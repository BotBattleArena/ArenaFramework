.PHONY: build test clean run-example

# install dependencies
install:
	go mod download

# Build all executables
build:
	go build ./...

# Build example binaries
build-examples: build-go-examples

build-go-examples:
	go build -o bin/topdownshooter.exe ./cmd/topdownshooter
	go build -o bin/shooterbot.exe ./cmd/shooterbot
	go build -o bin/hunterbot.exe ./cmd/hunterbot

# Run example: build randombot, set up inputs dir, run tictactoe
run-example: build-examples
	@if not exist inputs mkdir inputs
	@copy /y bin\shooterbot.exe inputs\1shooterbot-go.exe >nul
	@copy /y bin\hunterbot.exe inputs\2hunterbot-go.exe >nul
	bin\topdownshooter.exe

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	@if exist bin rmdir /s /q bin
	@if exist inputs rmdir /s /q inputs
	@if exist build rmdir /s /q build
	@if exist dist rmdir /s /q dist
	@if exist *.spec del /q *.spec
