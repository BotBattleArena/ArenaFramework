# Bot-Entwickler Guide

Diese Anleitung zeigt, wie man einen Input (Bot) für das ArenaFramework schreibt.

## Voraussetzungen

- Du bekommst die Datei `proto/arena/v1/arena.proto` vom Spieleentwickler.
- Du brauchst einen Protobuf-Compiler für deine Sprache.

## Protokoll

### Wire-Format

Jede Nachricht hat ein **Length-Prefix**:

```
[4 Bytes: Länge als uint32, Big-Endian][N Bytes: Protobuf-Payload]
```

### Nachrichten-Typen

**Vom Spiel empfangen (`ServerMessage`)**:

| type     | Bedeutung                          |
|----------|------------------------------------|
| `start`  | Spiel startet, `axes` enthält die verfügbaren Achsen |
| `state`  | Neuer Spielzustand in `state` (Bytes) |
| `end`    | Spiel ist vorbei                   |

**An das Spiel senden (`InputMessage`)**:

| Feld   | Typ                  | Bedeutung                    |
|--------|----------------------|------------------------------|
| `axes` | `map<string, float>` | Achsen-Werte von -1.0 bis 1.0 |

## Ablauf

1. **`start`** empfangen → Achsen-Namen merken
2. **`state`** empfangen → Spielzustand auswerten
3. **`InputMessage`** zurücksenden → Achsen-Werte setzen
4. Wiederhole 2-3 bis **`end`** kommt

## Beispiel: Go

```go
package main

import (
    "encoding/binary"
    "os"
    "math/rand"
    pb "path/to/generated/arena"
    "google.golang.org/protobuf/proto"
)

func main() {
    var axes []string
    for {
        var length uint32
        binary.Read(os.Stdin, binary.BigEndian, &length)

        buf := make([]byte, length)
        os.Stdin.Read(buf)

        msg := &pb.ServerMessage{}
        proto.Unmarshal(buf, msg)

        switch msg.Type {
        case "start":
            for _, ax := range msg.Axes {
                axes = append(axes, ax.Name)
            }
        case "state":
            resp := &pb.InputMessage{Axes: make(map[string]float32)}
            for _, name := range axes {
                resp.Axes[name] = rand.Float32()*2 - 1
            }
            out, _ := proto.Marshal(resp)
            binary.Write(os.Stdout, binary.BigEndian, uint32(len(out)))
            os.Stdout.Write(out)
        case "end":
            return
        }
    }
}
```

## Beispiel: Python

```python
import struct, sys
import arena_pb2  # generiert aus arena.proto mit: protoc --python_out=. arena.proto

axes = []

while True:
    raw_len = sys.stdin.buffer.read(4)
    if not raw_len:
        break
    length = struct.unpack(">I", raw_len)[0]
    data = sys.stdin.buffer.read(length)

    msg = arena_pb2.ServerMessage()
    msg.ParseFromString(data)

    if msg.type == "start":
        axes = [ax.name for ax in msg.axes]
    elif msg.type == "state":
        import random
        resp = arena_pb2.InputMessage()
        for name in axes:
            resp.axes[name] = random.uniform(-1, 1)
        out = resp.SerializeToString()
        sys.stdout.buffer.write(struct.pack(">I", len(out)))
        sys.stdout.buffer.write(out)
        sys.stdout.buffer.flush()
    elif msg.type == "end":
        break
```
