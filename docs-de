# Edigo: Kollaborativer Texteditor mit CRDT-basierter Konsistenzsicherung

## 1. Systemarchitektur

### 1.1 Überblick

Das System implementiert einen verteilten, kollaborativen Texteditor basierend auf dem Conflict-free Replicated Data Type (CRDT) Paradigma. Die Architektur folgt einem modularen Ansatz mit folgenden Hauptkomponenten:

- CRDT-Modul (crdt.go)
- Netzwerkmodul (network.go)
- Editor-Kern (editor.go)
- UI-Controller (ui.go)
- Syntax-Highlighter (highlighter.go)
- Theming-Engine (theme.go)

### 1.2 Technologie-Stack

- Programmiersprache: Go 1.16+
- UI-Framework: Bubble Tea (github.com/charmbracelet/bubbletea)
- Netzwerkprotokolle: UDP (Sitzungserkennung), TCP (Datenübertragung)

## 2. CRDT-Implementierung

### 2.1 Datenstruktur

Die CRDT-Implementierung basiert auf einer modifizierten Version des RGA (Replicated Growable Array) Algorithmus.

```go
type Element struct {
    ID        string
    Character rune
    Tombstone bool
}

type RGA struct {
    Elements       []Element
    Site           string
    Clock          int
    CursorPosition int
    RemoteCursors  map[string]int
    Checksum       uint32
}
```

### 2.2 Operationen

#### 2.2.1 Einfügen

```go
func (rga *RGA) LocalInsert(char rune) Operation {
    id := rga.generateID()
    newElement := Element{ID: id, Character: char, Tombstone: false}
    // Einfügelogik...
    rga.updateChecksum()
    return Operation{Type: Insert, ID: id, Character: char, Position: rga.CursorPosition}
}
```

#### 2.2.2 Löschen

```go
func (rga *RGA) LocalDelete() Operation {
    if rga.CursorPosition > 0 {
        rga.MoveCursorLeft()
        rga.Elements[rga.CursorPosition].Tombstone = true
        // Löschlogik...
        rga.updateChecksum()
        return Operation{Type: Delete, ID: rga.Elements[rga.CursorPosition].ID, Position: rga.CursorPosition}
    }
    return Operation{}
}
```

### 2.3 Konfliktauflösung

Die Konfliktauflösung basiert auf der totalen Ordnung der Element-IDs, die durch eine Kombination aus Standort-ID, logischer Uhr und Zufallswert erzeugt werden:

```go
func (rga *RGA) generateID() string {
    rga.Clock++
    return fmt.Sprintf("%s-%d-%d", rga.Site, rga.Clock, rand.Intn(1000))
}
```

### 2.4 Integritätsprüfung

Zur Sicherstellung der Datenintegrität wird ein CRC32-Checksum-Mechanismus implementiert:

```go
func (rga *RGA) updateChecksum() {
    data := []byte(rga.GetText())
    rga.Checksum = crc32.ChecksumIEEE(data)
}

func (rga *RGA) VerifyIntegrity() bool {
    currentChecksum := rga.Checksum
    rga.updateChecksum()
    return currentChecksum == rga.Checksum
}
```

## 3. Netzwerkprotokoll

### 3.1 Sitzungserkennung (UDP)

```go
func (network *Network) ListenForBroadcasts() {
    addr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":"+strconv.Itoa(network.UdpPort))
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    buffer := make([]byte, 1024)
    for {
        n, remoteAddr, _ := conn.ReadFromUDP(buffer)
        message := string(buffer[:n])
        // Verarbeitung der Broadcast-Nachricht...
    }
}
```

### 3.2 Datenübertragung (TCP)

```go
func (network *Network) SendOperation(op crdt.Operation, conn net.Conn) {
    bin_buf := new(bytes.Buffer)
    gobobj := gob.NewEncoder(bin_buf)
    gobobj.Encode(op)
    conn.Write(bin_buf.Bytes())
}
```

### 3.3 Serialisierung

Für die Serialisierung der Operationen und des RGA-Zustands wird das Gob-Protokoll verwendet:

```go
func SendInitRGA(rga crdt.RGA, conn net.Conn) {
    bin_buf := new(bytes.Buffer)
    gobobj := gob.NewEncoder(bin_buf)
    gobobj.Encode(rga)
    
    size := int64(bin_buf.Len())
    binary.Write(conn, binary.BigEndian, size)
    conn.Write(bin_buf.Bytes())
}
```

## 4. Syntax-Highlighting

### 4.1 Token-Definition

```go
type Token struct {
    Color  lipgloss.Style
    Value string
}

type Rule struct {
    Pattern *regexp.Regexp
    Color  lipgloss.Style
    Index   int
}
```

### 4.2 Lexer-Implementierung

```go
func (sd *SyntaxDefinition) LineLexer(line string) []Token {
    var tokens []Token
    remaining := line

    for len(remaining) > 0 {
        matched := false
        for _, rule := range sd.Rules {
            match := rule.Pattern.FindStringSubmatch(remaining)
            if len(match) > rule.Index {
                if rule.Pattern.FindIndex([]byte(remaining))[0] != 0 {continue}
                tokens = append(tokens, Token{rule.Color, match[rule.Index]})
                matched = true
                break
            }
        }
        // Weitere Tokenisierungslogik...
    }
    return tokens
}
```

### 4.3 Rendering

```go
func (sd *SyntaxDefinition) EmiteColorText(input string, output string) string {
    var result strings.Builder
    tokens := sd.LineLexer(input)
    remainingText := output

    for len(remainingText) > 0 {
        var color lipgloss.Style
        longestMatch := ""
        // Finde das längste passende Token...
        result.WriteString(sd.colorizeText(longestMatch, color))
        remainingText = remainingText[len(longestMatch):]
    }

    return result.String()
}
```

## 5. Leistungsoptimierungen

### 5.1 Effizientes Rendering

Um die Renderingperformanz bei großen Dokumenten zu optimieren, wird eine Virtualisierungstechnik implementiert:

```go
func (e *Editor) RenderContent() string {
    // ...
    totalLines := e.Viewport.Height - 2
    for i := 0; i < totalLines; i++ {
        if i < len(lines) {
            lineNumber := fmt.Sprintf("%*d", lineNumberWidth, i+1)
            renderedLineNumber := e.Theme.RenderLineNumber(lineNumber, lineNumberWidth)
            renderedLine := e.renderLineWithCursors(line, i)
            output.WriteString(renderedLineNumber + renderedLine + "\n")
        } else {
            // Render empty line...
        }
    }
    // ...
}
```

### 5.2 Optimierte Cursor-Bewegungen

```go
func (rga *RGA) MoveCursorRight() {
    for rga.CursorPosition < len(rga.Elements) {
        rga.CursorPosition++
        if rga.CursorPosition >= len(rga.Elements) {
            return
        }
        if !rga.Elements[rga.CursorPosition].Tombstone {
            break
        }
    }
}
```

## 6. Zukünftige Optimierungen und Erweiterungen
1. Implementierung eines Garbage Collection Mechanismus für gelöschte Elemente (Tombstones)
2. Integration eines verteilten Undo/Redo-Systems basierend auf Operation Transformation
3. Erweiterung des CRDT-Modells zur Unterstützung von strukturierten Daten (z.B. für Rich-Text-Formatierung)
4. Implementierung von Kompressionsalgorithmen für die Netzwerkübertragung zur Reduzierung der Bandbreitennutzung
5. Entwicklung eines Plugin-Systems zur einfachen Erweiterung der Editor-Funktionalität
