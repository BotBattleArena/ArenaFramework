// tictactoe is an example game that uses the arena framework.
// Two input processes play Tic-Tac-Toe on a 3x3 board.
// Each input has one axis: "position" (value 0-8 mapped from [-1,1]).
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/BotBattleArena/ArenaFramework/pkg/arena"
)

// Board represents a 3x3 Tic-Tac-Toe board.
// 0 = empty, 1 = player X, 2 = player O
type Board [9]int

// GameState is sent to the bots as JSON-encoded state bytes.
type GameState struct {
	Board         Board  `json:"board"`
	YourSymbol    int    `json:"your_symbol"`    // 1 or 2
	CurrentPlayer int    `json:"current_player"` // 1 or 2
	Message       string `json:"message,omitempty"`
}

func main() {
	fmt.Println("=== Tic-Tac-Toe ===")
	fmt.Println("Starting game with inputs from ./inputs/ directory...")
	fmt.Println()

	a, err := arena.New(arena.Config{
		InputDir:      "./inputs",
		ActionTimeout: 5 * time.Second,
		Axes: []arena.Axis{
			{Name: "position", Value: 0},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create arena: %v", err)
	}

	a.OnConnect(func(p arena.Player) {
		fmt.Printf("  Player connected: %s\n", p.ID)
	})

	a.OnDisconnect(func(p arena.Player, err error) {
		fmt.Printf("  Player disconnected: %s (error: %v)\n", p.ID, err)
	})

	if err := a.Start(); err != nil {
		log.Fatalf("Failed to start arena: %v", err)
	}
	defer a.Stop()

	players := a.Players()
	if len(players) < 2 {
		log.Fatalf("Need at least 2 inputs, got %d", len(players))
	}

	player1 := players[0].ID
	player2 := players[1].ID
	symbols := map[string]int{player1: 1, player2: 2}
	symbolNames := map[int]string{1: "X", 2: "O"}

	fmt.Printf("  %s plays as X\n", player1)
	fmt.Printf("  %s plays as O\n", player2)
	fmt.Println()

	var board Board
	currentPlayer := player1

	for turn := 0; turn < 9; turn++ {
		symbol := symbols[currentPlayer]

		// Send state and request axes from all players
		stateForCurrent, _ := json.Marshal(GameState{
			Board:         board,
			YourSymbol:    symbol,
			CurrentPlayer: symbol,
		})
		responses := a.RequestAxes(stateForCurrent, 5*time.Second)

		// Convert axis value [-1, 1] to board position [0, 8]
		axes := responses[currentPlayer]
		posFloat := axes["position"]
		pos := int(math.Round(float64(posFloat+1) / 2 * 8))
		if pos < 0 {
			pos = 0
		}
		if pos > 8 {
			pos = 8
		}

		// Find first valid position starting from chosen position
		placed := false
		for i := 0; i < 9; i++ {
			tryPos := (pos + i) % 9
			if board[tryPos] == 0 {
				board[tryPos] = symbol
				fmt.Printf("Turn %d: %s (%s) places at position %d\n", turn+1, currentPlayer, symbolNames[symbol], tryPos)
				placed = true
				break
			}
		}

		if !placed {
			fmt.Println("No valid moves left!")
			break
		}

		printBoard(board, symbolNames)

		// Check for winner
		if winner := checkWinner(board); winner != 0 {
			fmt.Printf("\n🎉 %s (%s) wins!\n", currentPlayer, symbolNames[winner])
			return
		}

		// Switch player
		if currentPlayer == player1 {
			currentPlayer = player2
		} else {
			currentPlayer = player1
		}
	}

	fmt.Println("\n🤝 It's a draw!")
}

func printBoard(board Board, symbols map[int]string) {
	chars := func(v int) string {
		if s, ok := symbols[v]; ok {
			return s
		}
		return "."
	}
	fmt.Println()
	for row := 0; row < 3; row++ {
		fmt.Printf("  %s | %s | %s\n",
			chars(board[row*3]),
			chars(board[row*3+1]),
			chars(board[row*3+2]))
		if row < 2 {
			fmt.Println("  ---------")
		}
	}
	fmt.Println()
}

func checkWinner(board Board) int {
	lines := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	for _, line := range lines {
		if board[line[0]] != 0 &&
			board[line[0]] == board[line[1]] &&
			board[line[1]] == board[line[2]] {
			return board[line[0]]
		}
	}
	return 0
}
