.PHONY: build test clean run-example

# install dependencies
install:
	go mod download

# Build all executables
build:
	go build ./...

# Build example binaries
build-examples:
	go build -o bin/tictactoe.exe ./cmd/tictactoe
	go build -o bin/randombot.exe ./cmd/randombot

# Run example: build randombot, set up inputs dir, run tictactoe
run-example: build-examples
	@if not exist inputs mkdir inputs
	@copy /y bin\randombot.exe inputs\player1.exe >nul
	@copy /y bin\randombot.exe inputs\player2.exe >nul
	bin\tictactoe.exe

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	@if exist bin rmdir /s /q bin
	@if exist inputs rmdir /s /q inputs
	@if exist gen rmdir /s /q gen
