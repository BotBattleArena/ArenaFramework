.PHONY: proto build test clean run-example

# Generate Go code from proto files
proto:
	protoc --go_out=./gen --go_opt=paths=source_relative \
		--proto_path=proto \
		proto/arena/v1/arena.proto

# Build all executables
build:
	go build ./...

# Build example binaries
build-examples:
	go build -o bin/tictactoe.exe ./cmd/tictactoe
	go build -o bin/randombot.exe ./cmd/randombot

# Run example: build randombot, set up inputs dir, run tictactoe
run-example: build-examples
	@mkdir -p inputs
	@cp bin/randombot.exe inputs/player1.exe
	@cp bin/randombot.exe inputs/player2.exe
	bin/tictactoe.exe

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/ inputs/
