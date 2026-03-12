package main

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"os"

	pb "github.com/BotBattleArena/ArenaFramework/gen/go/arena/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	log.SetOutput(os.Stderr) // Log to stderr so it doesn't interfere with the protocol
	log.Println("randombot: started")

	var axisNames []string

	for {
		// Read length prefix (4 bytes, big-endian)
		var length uint32
		if err := binary.Read(os.Stdin, binary.BigEndian, &length); err != nil {
			log.Printf("randombot: read error: %v", err)
			return
		}

		// Read protobuf payload
		buf := make([]byte, length)
		if _, err := io.ReadFull(os.Stdin, buf); err != nil {
			log.Printf("randombot: read payload error: %v", err)
			return
		}

		// Decode ServerMessage
		msg := &pb.ServerMessage{}
		if err := proto.Unmarshal(buf, msg); err != nil {
			log.Printf("randombot: unmarshal error: %v", err)
			continue
		}

		switch msg.Type {
		case "start":
			// Store axis names from the game
			axisNames = make([]string, len(msg.Axes))
			for i, ax := range msg.Axes {
				axisNames[i] = ax.Name
				log.Printf("randombot: axis registered: %s (default: %.2f)", ax.Name, ax.Value)
			}

		case "state":
			// Respond with random axis values between -1.0 and 1.0
			response := &pb.InputMessage{
				Axes: make(map[string]float32, len(axisNames)),
			}
			for _, name := range axisNames {
				response.Axes[name] = rand.Float32()*2 - 1 // [-1.0, 1.0]
			}

			// Encode and send
			out, err := proto.Marshal(response)
			if err != nil {
				log.Printf("randombot: marshal error: %v", err)
				continue
			}
			if err := binary.Write(os.Stdout, binary.BigEndian, uint32(len(out))); err != nil {
				log.Printf("randombot: write length error: %v", err)
				return
			}
			if _, err := os.Stdout.Write(out); err != nil {
				log.Printf("randombot: write payload error: %v", err)
				return
			}

		case "end":
			log.Println("randombot: game ended")
			return
		}
	}
}
