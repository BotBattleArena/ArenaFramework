# Bot-Entwickler Guide

Diese Anleitung zeigt, wie man einen Input (Bot) für das ArenaFramework schreibt.

## Voraussetzungen

- Du musst JSON in deiner Programmiersprache verarbeiten können.
- Dein Bot kommuniziert über **stdin** (Eingabe vom Spiel) und **stdout** (Ausgabe ans Spiel).

## Protokoll

### Wire-Format

Das Framework nutzt **NDJSON** (Newline-Delimited JSON). Jede Nachricht ist ein einzelnes JSON-Objekt in einer Zeile, gefolgt von einem Newline-Charakter (`\n`).

- **Lesen**: Lies von stdin bis zum nächsten `\n`, parset dann den String als JSON.
- **Schreiben**: Serialisiere dein JSON-Objekt zu einem String und schreibe ihn gefolgt von `\n` auf stdout. **Wichtig**: Nutze `flush`, um sicherzustellen, dass die Nachricht sofort gesendet wird.

### Nachrichten-Typen

Das Schema der Nachrichten findest du in `docs/bot-schema.json`.

**Vom Spiel empfangen (`ServerMessage`)**:

| Feld   | Typ      | Bedeutung                                   |
|--------|----------|---------------------------------------------|
| `type` | `string` | `start` (Initial), `state` (Update), `end` (Ende) |
| `axes` | `array`  | Liste der verfügbaren Achsen (nur bei `start`) |
| `state`| `object` | Spiel-spezifischer Zustand (nur bei `state`)   |

**An das Spiel senden (`InputMessage`)**:

| Feld   | Typ                  | Bedeutung                    |
|--------|----------------------|------------------------------|
| `axes` | `map<string, float>` | Achsen-Werte von -1.0 bis 1.0 |

## Ablauf

1. **`start`** empfangen → Achsen-Namen aus der `axes`-Liste merken.
2. **`state`** empfangen → Den aktuellen Spielzustand auswerten.
3. **`InputMessage`** zurücksenden → Ein JSON-Objekt mit den gewünschten `axes`-Werten schreiben.
4. Wiederhole 2-3 bis **`end`** kommt oder die Pipe geschlossen wird.

## Beispiel: Go

```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
)

type ServerMessage struct {
    Type string `json:"type"`
    Axes []struct {
        Name string `json:"name"`
    } `json:"axes"`
}

type InputMessage struct {
    Axes map[string]float32 `json:"axes"`
}

func main() {
    var axisNames []string
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        var msg ServerMessage
        json.Unmarshal(scanner.Bytes(), &msg)

        switch msg.Type {
        case "start":
            for _, ax := range msg.Axes {
                axisNames = append(axisNames, ax.Name)
            }
        case "state":
            resp := InputMessage{Axes: make(map[string]float32)}
            for _, name := range axisNames {
                resp.Axes[name] = rand.Float32()*2 - 1
            }
            out, _ := json.Marshal(resp)
            fmt.Println(string(out))
        case "end":
            return
        }
    }
}
```

## Beispiel: Python

```python
import sys
import json
import random

axis_names = []

# sys.stdin ist ein Iterator, der Zeile für Zeile liest
for line in sys.stdin:
    msg = json.loads(line)
    
    if msg["type"] == "start":
        axis_names = [ax["name"] for ax in msg["axes"]]
        
    elif msg["type"] == "state":
        response = {
            "axes": {name: random.uniform(-1, 1) for name in axis_names}
        }
        # json.dumps + print erzeugt NDJSON (+ \n)
        # flush=True ist wichtig für Echtzeit-Kommunikation
        print(json.dumps(response), flush=True)
        
    elif msg["type"] == "end":
        break
```
