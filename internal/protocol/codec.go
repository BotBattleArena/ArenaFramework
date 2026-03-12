package protocol

import (
	"fmt"
	"io"

	pb "github.com/BotBattleArena/ArenaFramework/gen/go/arena/v1"
	"google.golang.org/protobuf/proto"
)

// WriteServerMessage encodes and writes a ServerMessage as a length-prefixed frame.
func WriteServerMessage(w io.Writer, msg *pb.ServerMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal server message: %w", err)
	}
	return WriteFrame(w, data)
}

// ReadServerMessage reads a length-prefixed frame and decodes it as a ServerMessage.
func ReadServerMessage(r io.Reader) (*pb.ServerMessage, error) {
	data, err := ReadFrame(r)
	if err != nil {
		return nil, err
	}
	msg := &pb.ServerMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("unmarshal server message: %w", err)
	}
	return msg, nil
}

// WriteInputMessage encodes and writes an InputMessage as a length-prefixed frame.
func WriteInputMessage(w io.Writer, msg *pb.InputMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal input message: %w", err)
	}
	return WriteFrame(w, data)
}

// ReadInputMessage reads a length-prefixed frame and decodes it as an InputMessage.
func ReadInputMessage(r io.Reader) (*pb.InputMessage, error) {
	data, err := ReadFrame(r)
	if err != nil {
		return nil, err
	}
	msg := &pb.InputMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("unmarshal input message: %w", err)
	}
	return msg, nil
}
