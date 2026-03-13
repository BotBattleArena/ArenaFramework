package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/BotBattleArena/ArenaFramework/pkg/arena"
)

type BotFrame struct {
	Players map[string]BotPlayer `json:"players"`
	Bullets []BotBullet          `json:"bullets"`
	MapW    float64              `json:"map_w"`
	MapH    float64              `json:"map_h"`
	Tick    int                  `json:"tick"`
	Left    int                  `json:"time_left"`
}

type BotPlayer struct {
	ID      string  `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	HP      int     `json:"hp"`
	Kills   int     `json:"kills"`
	Deaths  int     `json:"deaths"`
	Alive   bool    `json:"alive"`
	ShootCD int     `json:"shoot_cd"`
	DashCD  int     `json:"dash_cd"`
	AimX    float64 `json:"aim_x"`
	AimY    float64 `json:"aim_y"`
	Color   string  `json:"color"`
}

type BotBullet struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	DX    float64 `json:"dx"`
	DY    float64 `json:"dy"`
	Owner string  `json:"owner"`
}

func getMyID() string {
	name := filepath.Base(os.Args[0])
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return name
}

func main() {
	log.SetOutput(os.Stderr)
	myID := getMyID()
	log.Printf("hunterbot: started as %s", myID)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*256), 1024*256)

	for scanner.Scan() {
		var msg arena.ServerMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			log.Printf("hunterbot: parse error: %v", err)
			continue
		}

		switch msg.Type {
		case "start":
			log.Println("hunterbot: game started")

		case "state":
			var frame BotFrame
			if err := json.Unmarshal(msg.State, &frame); err != nil {
				log.Printf("hunterbot: state error: %v", err)
				continue
			}

			me, ok := frame.Players[myID]
			if !ok || !me.Alive {
				resp := arena.InputMessage{Axes: map[string]float32{
					"move_x": 0, "move_y": 0,
					"aim_x": 0, "aim_y": 1,
					"shoot": 0, "dash": 0,
				}}
				out, _ := json.Marshal(resp)
				fmt.Println(string(out))
				continue
			}

			// AGGRESSIVE HUNTING LOGIC
			closest := math.MaxFloat64
			var target *BotPlayer
			for id, p := range frame.Players {
				if id == myID || !p.Alive {
					continue
				}
				dx, dy := p.X-me.X, p.Y-me.Y
				d := math.Sqrt(dx*dx + dy*dy)
				if d < closest {
					closest = d
					target = &p
				}
			}

			var mx, my, ax, ay float32
			shoot := float32(0)
			dash := float32(0)

			if target != nil {
				dx, dy := target.X-me.X, target.Y-me.Y
				dist := math.Sqrt(dx*dx + dy*dy)
				
				// Move towards target
				mx = float32(dx / dist)
				my = float32(dy / dist)

				// Aim at target
				ax, ay = mx, my

				// Shoot if in range
				if dist < 600 && me.ShootCD == 0 {
					shoot = 1
				}

				// Dash to close the gap if far away
				if dist > 300 && me.DashCD == 0 && rand.Float64() < 0.1 {
					dash = 1
				}
			} else {
				// No target? Move towards center
				cx, cy := frame.MapW/2-me.X, frame.MapH/2-me.Y
				d := math.Sqrt(cx*cx + cy*cy)
				if d > 10 {
					mx, my = float32(cx/d), float32(cy/d)
				}
				ax, ay = 0, 1
			}

			// EVASION LOGIC (Dodge incoming bullets)
			for _, b := range frame.Bullets {
				if b.Owner == myID {
					continue
				}
				dx, dy := b.X-me.X, b.Y-me.Y
				d := math.Sqrt(dx*dx + dy*dy)
				if d < 150 {
					// Is bullet moving towards me?
					dot := dx*b.DX + dy*b.DY
					if dot < 0 {
						// Perpendicular dodge
						evadeX, evadeY := -b.DY, b.DX
						if rand.Float64() < 0.5 {
							evadeX, evadeY = -evadeX, -evadeY
						}
						mx, my = float32(evadeX), float32(evadeY)
						// Use dash for emergency evasion if very close
						if d < 70 && me.DashCD == 0 {
							dash = 1
						}
						break
					}
				}
			}

			// Boundary check
			margin := 100.0
			if me.X < margin || me.X > frame.MapW-margin || me.Y < margin || me.Y > frame.MapH-margin {
				cx, cy := frame.MapW/2-me.X, frame.MapH/2-me.Y
				cd := math.Sqrt(cx*cx + cy*cy)
				mx = (mx + float32(cx/cd)*2) / 3
				my = (my + float32(cy/cd)*2) / 3
			}

			resp := arena.InputMessage{Axes: map[string]float32{
				"move_x": mx, "move_y": my,
				"aim_x": ax, "aim_y": ay,
				"shoot": shoot, "dash": dash,
			}}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))

		case "end":
			log.Println("hunterbot: game over")
			return
		}
	}
}
