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
Allerdings möchten wir transparent darüber informieren, dass die Einfügeoperation noch nicht vollständig den idealen CRDT-Prinzipien entspricht.
Die aktuelle Implementierung bietet bereits eine solide Grundlage für kollaboratives Editieren und bewältigt viele Szenarien erfolgreich. Dennoch sind wir uns bewusst, dass in bestimmten komplexen Situationen noch Verbesserungspotenzial besteht.

Um die CRDT-Implementierung zu vervollständigen, sind folgende Schritte erforderlich:

1. Verfeinerte Timestamp-Generierung: Implementierung eines präziseren Mechanismus zur Erzeugung eindeutiger Zeitstempel, der die kausale Beziehung zwischen Operationen besser abbildet.
2. Erweiterte Konflikterkennung: Entwicklung fortschrittlicherer Algorithmen zur Erkennung und Auflösung von Konflikten bei gleichzeitigen Einfügeoperationen an derselben Position.
3. Optimierte Datenstruktur: Überarbeitung der zugrundeliegenden Datenstruktur, um eine effizientere Verwaltung und Zusammenführung von Einfügeoperationen zu ermöglichen.
4. Verbesserte Netzwerksynchronisation: Implementierung eines robusteren Protokolls für die Übertragung und Synchronisation von CRDT-Operationen zwischen den Clients.

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

[Vorherige Inhalte bleiben unverändert]

## 7. Systembedienung und Anwendungsfälle

Dieses Kapitel erläutert die praktische Nutzung des kollaborativen Texteditors und beschreibt typische Anwendungsfälle.

### 7.1 Systemstart und Konfiguration

#### 7.1.1 Kompilierung und Ausführung

```bash
go build -o editor
./editor <Dateipfad>
```

#### 7.1.2 Konfigurationsparameter

Der Editor akzeptiert folgende Kommandozeilenargumente:

- `-port=<Portnummer>`: Spezifiziert den zu verwendenden Netzwerkport (Standard: 12346)
- `-theme=<Themenname>`: Wählt ein vordefiniertes Farbschema (Standard: "default")

Beispiel:
```bash
./editor -port=12350 -theme=dark mydocument.txt
```

### 7.2 Benutzeroberfläche und Navigation

Die Benutzeroberfläche ist in drei Hauptbereiche unterteilt:

1. Headerbereich: Zeigt Dateinamen und Verbindungsstatus
2. Editorbereich: Hauptbereich für die Textbearbeitung
3. Footerbereich: Zeigt Statusinformationen und Fehlermeldungen

Navigation erfolgt primär über die Pfeiltasten:

- ↑/↓: Bewegt den Cursor vertikal
- ←/→: Bewegt den Cursor horizontal
- Strg+←/→: Bewegt den Cursor wortweise

### 7.3 Kollaborative Funktionen

#### 7.3.1 Sitzungserstellung

```go
func (network *Network) BroadcastSession(rga *crdt.RGA) {
    // ...
    go func() {
        for {
            // Broadcast session information
            message := []byte(fmt.Sprintf("SESSION|%s|%d|%d|%s|%s", sessionName, port, network.UdpPort, network.HostFilePath, network.HostFileExt))
            // Send broadcast...
        }
    }()
    // ...
}
```

Um eine neue Sitzung zu erstellen:
1. Öffnen Sie das Menü mit `Esc`
2. Wählen Sie "Create Session" → "Create Public Session"

#### 7.3.2 Sitzungsbeitritt

```go
func (network *Network) JoinSession(sessionName string) crdt.RGA {
    // ...
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", session.IP, session.Port))
    // ...
    // Receive and apply initial RGA state
    // ...
    return *tmpstruct
}
```

Um einer Sitzung beizutreten:
1. Öffnen Sie das Menü mit `Esc`
2. Wählen Sie "Join Session"
3. Wählen Sie die gewünschte Sitzung aus der Liste

### 7.4 Textbearbeitung und CRDT-Operationen

#### 7.4.1 Texteinfügung

```go
func (ih *InputHandler) HandleKeyMsg(msg tea.KeyMsg) {
    // ...
    case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
        ih.Editor.InsertCharacter('\n')
    default:
        if len(msg.String()) == 1 { // Only handle single characters
            ih.Editor.InsertCharacter(rune(msg.String()[0]))
        }
    // ...
}
```

Texteinfügung erfolgt durch direkte Tastatureingabe. Jede Einfügung erzeugt eine CRDT-Operation, die lokal angewendet und an verbundene Clients gesendet wird.

#### 7.4.2 Textlöschung

```go
func (ih *InputHandler) HandleKeyMsg(msg tea.KeyMsg) {
    // ...
    case key.Matches(msg, key.NewBinding(key.WithKeys("backspace", "ctrl+h"))):
        ih.Editor.DeleteCharacterBeforeCursor()
    // ...
}
```

Löschungen werden durch die Rücktaste ausgelöst und erzeugen ebenfalls CRDT-Operationen.

### 7.5 Dateioperationen

#### 7.5.1 Speichern

```go
func (m *UIModel) saveFile() {
    // ...
    content := m.Editor.RenderDocumentWithoutLineNumbers()
    err := os.WriteFile(m.Editor.FilePath, []byte(content), 0644)
    // ...
}
```

Zum Speichern:
1. Drücken Sie `Strg+S`, oder
2. Öffnen Sie das Menü mit `Esc` und wählen Sie "Save"

#### 7.5.2 Laden

Das Laden einer Datei erfolgt beim Programmstart durch Angabe des Dateipfads als Kommandozeilenargument.

### 7.6 Fortgeschrittene Funktionen

#### 7.6.1 Syntax-Highlighting

Das Syntax-Highlighting wird automatisch basierend auf der Dateierweiterung aktiviert:

```go
func GetSyntaxDefiniton(fileEnd string) *SyntaxDefinition {
    switch (fileEnd) {
    case ".py":
        return NewPythonSyntaxDefinition()
    case ".js":
        return NewJavaScriptSyntaxDefinition()
    case ".html":
        return NewHTMLSyntaxDefinition()
    }
    return NewDefaultSyntaxDefinition()
}
```

#### 7.6.2 Theming

Themes können über das Kommandozeilenargument `-theme` ausgewählt werden. Die Theming-Engine ermöglicht die Anpassung aller UI-Elemente:

```go
type Theme struct {
    BaseStyle             lipgloss.Style
    HeaderStyle           lipgloss.Style
    FooterStyle           lipgloss.Style
    LineNumberStyle       lipgloss.Style
    CursorStyle           lipgloss.Style
    // ...
}
```

### 7.7 Fehlerbehebung und Debugging

#### 7.7.1 Verbindungsprobleme

Bei Verbindungsproblemen:
1. Überprüfen Sie die Netzwerkeinstellungen
2. Stellen Sie sicher, dass der spezifizierte Port verfügbar ist
3. Versuchen Sie, der Sitzung erneut beizutreten

#### 7.7.2 Konsistenzprüfung

Der Editor führt regelmäßige Konsistenzprüfungen durch:

```go
func (rga *RGA) VerifyIntegrity() bool {
    currentChecksum := rga.Checksum
    rga.updateChecksum()
    return currentChecksum == rga.Checksum
}
```
Bei Inkonsistenzen wird eine Warnmeldung angezeigt, und es wird empfohlen, die Sitzung neu zu verbinden.
