package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/BotBattleArena/ArenaFramework/pkg/arena"
)

func main() {
	log.SetOutput(os.Stderr) // Log to stderr so it doesn't interfere with the protocol
	log.Println("randombot: started")

	var axisNames []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var msg arena.ServerMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
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
			response := arena.InputMessage{
				Axes: make(map[string]float32, len(axisNames)),
			}
			for _, name := range axisNames {
				response.Axes[name] = rand.Float32()*2 - 1 // [-1.0, 1.0]
			}

			// Encode and send
			out, err := json.Marshal(response)
			if err != nil {
				log.Printf("randombot: marshal error: %v", err)
				continue
			}
			fmt.Println(string(out)) // Println adds the newline delimiter

		case "end":
			log.Println("randombot: game ended")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("randombot: scanner error: %v", err)
	}
}
