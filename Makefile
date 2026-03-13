.PHONY: build test clean run-example

# install dependencies
install:
	go mod download

# Build all executables
build:
	go build ./...

# Build example binaries
build-examples: build-go-examples build-python-bot

build-go-examples:
	go build -o bin/tictactoe.exe ./cmd/tictactoe
	go build -o bin/randombot.exe ./cmd/randombot

build-python-bot:
	@if not exist bin mkdir bin
	pyinstaller --onefile --name randombot-python cmd/randombot-python/main.py
	@move /y dist\randombot-python.exe bin\ >nul
	@rmdir /s /q build
	@rmdir /s /q dist
	@del /q randombot-python.spec

# Run example: build randombot, set up inputs dir, run tictactoe
run-example: build-examples
	@if not exist inputs mkdir inputs
	@copy /y bin\randombot.exe inputs\randombot-go.exe >nul
	@copy /y bin\randombot-python.exe inputs\randombot-python.exe >nul
	bin\tictactoe.exe

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
