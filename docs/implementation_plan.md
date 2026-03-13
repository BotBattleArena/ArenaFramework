# Architektur- und Implementierungsplan: ArenaFramework

Go-Library, die ein Spiel importiert um **Input-EXEs** (Bots, Spieler, KI, etc.) als Subprozesse zu managen. Inputs sind **Achsen mit Werten von -1 bis 1**. Kommunikation über **stdin/stdout mit JSON (NDJSON)**.

## Architektur

```
┌───────────────────────────────────────────────┐
│         Spiel-Executable (inkl. UI)           │
│   import "arena"                              │
│                                               │
│  ┌─────────────────────────────────────────┐  │
│  │          arena (Library)                │  │
│  │  Lädt alle EXEs aus InputDir            │  │
│  │  Goroutine pro Input-Prozess            │  │
│  └──────┬──────────────┬───────────────────┘  │
└─────────│──────────────│──────────────────────┘
          │stdin/stdout  │stdin/stdout
    ┌─────▼─────┐  ┌─────▼─────┐
    │ random.exe │  │ smart.exe │
    └───────────┘  └───────────┘

Input-Name = Dateiname der EXE (ohne Extension)
```

---

## Config

```go
type Config struct {
    InputDir      string        // Verzeichnis mit Input-EXEs
    Axes          []Axis        // Welche Achsen das Spiel hat
    ActionTimeout time.Duration // Max Wartezeit auf Antwort
}

type Axis struct {
    Name  string  // z.B. "move_x", "shoot"
    Value float32 // Default-Wert (normalerweise 0.0)
}
```

Das Spiel gibt nur ein Verzeichnis an. Die Library findet alle `.exe`-Dateien darin automatisch. Der Input-Name wird vom Dateinamen abgeleitet (`random.exe` → ID `"random"`).

---

## API

### Lifecycle

| Methode | Beschreibung |
|---|---|
| `arena.New(cfg Config) *Arena` | Neue Arena |
| `Start() error` | Startet alle Input-Prozesse aus `InputDir` |
| `Stop()` | Beendet alle Prozesse |

### Kommunikation (Spiel → Input)

| Methode | Beschreibung |
|---|---|
| `SendState(state []byte)` | State an alle Inputs senden |
| `SendStateTo(inputID string, state []byte)` | State an einen Input senden |
| `RequestAxes(timeout time.Duration) map[string]map[string]float32` | State senden + auf Achsen-Werte aller Inputs warten |

### Events (Input → Spiel)

| Methode | Beschreibung |
|---|---|
| `OnAxes(handler func(p Player, axes map[string]float32))` | Achsen-Werte empfangen |
| `OnConnect(handler func(p Player))` | Input-Prozess bereit |
| `OnDisconnect(handler func(p Player, err error))` | Input-Prozess beendet/crashed |

### Abfragen

| Methode | Beschreibung |
|---|---|
| `Players() []Player` | Alle Inputs |
| `Player(id string) (Player, bool)` | Einzelnen Input |
| `IsRunning() bool` | Arena aktiv? |

---

## Player Struct

```go
type Player struct {
    ID     string       // Dateiname ohne Extension ("random")
    Status PlayerStatus // Connected, Disconnected, TimedOut
}
```

---

## Protokoll (sprachunabhängig, JSON-basiert)

**Wire-Format**: [NDJSON](https://github.com/ndjson/ndjson-spec) (Newline-Delimited JSON).

```json
// Beispiel: ServerMessage (Spiel -> Input)

// Spiel → Input
message ServerMessage {
    string type = 1;        // "state", "start", "end"
    bytes  state = 2;       // Spiel-spezifische Daten (opaque)
    repeated Axis axes = 3; // Achsen-Definitionen (nur bei "start")
}

message Axis {
    string name = 1;   // "move_x"
    float  value = 2;  // Default-Wert (0.0)
}

// Input → Spiel
message InputMessage {
    map<string, float> axes = 1; // {"move_x": 0.5, "shoot": 1.0}
}
```

### Input-Seite: Go (pur, kein Package)

```go
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        var msg ServerMessage
        json.Unmarshal(scanner.Bytes(), &msg)

        // Antwort senden
        resp := InputMessage{Axes: map[string]float32{"move_x": 0.5}}
        out, _ := json.Marshal(resp)
        fmt.Println(string(out)) // \n delimiter
    }
}
```

### Input-Seite: Python (pur, kein Package)

```python
import sys, json

for line in sys.stdin:
    msg = json.loads(line)
    
    # Antwort senden
    resp = {"axes": {"move_x": 0.5}}
    print(json.dumps(resp), flush=True)
```

---

ArenaFramework/
│
├── pkg/                                # Öffentliche Library (importierbar)
│   └── arena/
│       ├── arena.go                    # Arena struct, Config, öffentliche API
│       ├── options.go                  # Functional Options (WithTimeout, etc.)
│       └── types.go                    # Player, Axis, PlayerStatus
│
├── internal/                           # Private Implementierung (nicht importierbar)
│   ├── session/
│   │   ├── manager.go                  # SessionManager: Subprozesse starten/stoppen
│   │   └── process.go                  # Einzelner Input-Prozess + Goroutinen
│   └── protocol/
│       └── codec.go                    # NDJSON Encoder/Decoder
│
├── cmd/                                # Beispiel-Executables
│   ├── tictactoe/                      # Beispiel-Spiel (importiert pkg/arena)
│   │   └── main.go
│   └── randombot/                      # Beispiel-Input-EXE
│       └── main.go
│
├── docs/                               # Dokumentation
│   ├── bot-guide.md                    # Anleitung für Bot-Entwickler
│   └── bot-schema.json                 # JSON-Schema für das Protokoll
│
├── Makefile                            # build, test
├── go.mod
├── go.sum
└── README.md
```

| Verzeichnis | Zweck |
|---|---|
| `pkg/arena/` | Öffentliche API – das was Spiele importieren |
| `internal/` | Private Implementierung, nicht von außen importierbar |
| `cmd/` | Ausführbare Beispiele |
| `docs/` | Dokumentation (Guide, Schema) |

## Verification Plan

### Automated Tests
- `go test ./pkg/arena/... ./internal/...` – Protocol Roundtrip, Session mit Dummy-Input

### Manual Verification
- TicTacToe mit 2x RandomBot starten, Spielverlauf + Gewinner verifizieren
