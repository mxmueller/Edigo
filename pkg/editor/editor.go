package editor

import (
	"bytes"
	"edigo/pkg/crdt"
	"edigo/pkg/highlighter"
	"edigo/pkg/network"
	"edigo/pkg/theme"
	"encoding/gob"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
)

type RemoteChange struct{}

type CursorInfo struct {
	Position   int
	LastMove   time.Time
	Username   string
	ThemeIndex int
}

type Editor struct {
	RGA             *crdt.RGA
	Viewport        viewport.Model
	Network         *network.Network
	FilePath        string
	Update          chan struct{}
	NewConnection   chan net.Conn
	Error           string
	Theme           *theme.Theme
	SyntaxDef       highlighter.SyntaxDefinition
	LocalCursor     CursorInfo
	RemoteCursors   map[string]CursorInfo
	remoteCursorMu  sync.RWMutex
	updateTicker    *time.Ticker
	nextThemeIndex  int
	IsSharedSession bool
	guestCounter    int
}

func NewEditor(content string, filePath string, siteID string, theme *theme.Theme) *Editor {
	rga := crdt.NewRGA(siteID)
	for _, char := range content {
		rga.LocalInsert(char)
	}

	newConnection := make(chan net.Conn, 1)
	network := network.NewNetwork()
	network.NewConnection = newConnection

	fileType := filepath.Ext(filePath)

	editor := &Editor{
		RGA:           rga,
		Network:       network,
		NewConnection: newConnection,
		Theme:         theme,
		SyntaxDef:     *highlighter.GetSyntaxDefiniton(fileType),
		Viewport:      viewport.New(80, 24),
		LocalCursor: CursorInfo{
			Position:   0,
			LastMove:   time.Now(),
			Username:   "You",
			ThemeIndex: 0,
		},
		RemoteCursors:   make(map[string]CursorInfo),
		updateTicker:    time.NewTicker(100 * time.Millisecond),
		Update:          make(chan struct{}, 1),
		nextThemeIndex:  1,
		IsSharedSession: false,
		guestCounter:    0,
		FilePath:        filePath,
	}
	return editor
}

func (e *Editor) Stop() {
	e.updateTicker.Stop()
}

func (e *Editor) InsertCharacter(ch rune) {
	op := e.RGA.LocalInsert(ch)
	e.sendToRemote(op)
	e.updateLocalCursor()
}

func (e *Editor) DeleteCharacterBeforeCursor() {
	op := e.RGA.LocalDelete()
	e.sendToRemote(op)
	e.updateLocalCursor()
}

func (e *Editor) MoveCursorLeft() {
	e.RGA.MoveCursorLeft()
	e.updateLocalCursor()
}

func (e *Editor) MoveCursorRight() {
	e.RGA.MoveCursorRight()
	e.updateLocalCursor()
}

func (e *Editor) MoveCursorUp() {
	e.RGA.MoveCursorUp()
	e.updateLocalCursor()
}

func (e *Editor) MoveCursorDown() {
	e.RGA.MoveCursorDown()
	e.updateLocalCursor()
}

func (e *Editor) updateLocalCursor() {
	e.LocalCursor.Position = e.RGA.CursorPosition
	e.LocalCursor.LastMove = time.Now()
	e.SendCursorUpdate()
}

func (e *Editor) SendCursorUpdate() {
	go e.sendToRemote(crdt.Operation{Type: crdt.Move, ID: e.Network.ID, Character: 0, Position: e.LocalCursor.Position})
}

func (e *Editor) sendToRemote(op crdt.Operation) {
	if e.Network.IsHost {
		for _, conn := range e.Network.Clients {
			e.Network.SendOperation(op, conn)
		}
	} else if e.Network.Host != nil {
		e.Network.SendOperation(op, e.Network.Host)
	}
}

func (e *Editor) reciveInput(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 256)

		conn.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
		_, err := conn.Read(buf)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			if e.Network.IsHost {
				e.Network.RemoveClient(conn)
			}
			if e.Network.Host != nil {
				e.Network.HostClosedSession()
			}
			e.Update <- struct{}{}
			return
		}
		tmpbuff := bytes.NewBuffer(buf)
		incomingOp := new(crdt.Operation)

		gobobj := gob.NewDecoder(tmpbuff)
		gobobj.Decode(incomingOp)

		if incomingOp.Type == crdt.Move {
			e.updateRemoteCursor(incomingOp.ID, incomingOp.Position)
		} else {
			oldPosition := e.LocalCursor.Position
			e.RGA.ApplyOperation(*incomingOp)
			e.adjustCursorAfterRemoteOp(oldPosition, incomingOp)
		}

		e.Update <- struct{}{}

		if e.Network.IsHost {
			for _, sendConn := range e.Network.Clients {
				if sendConn == conn {
					continue
				}
				e.Network.SendOperation(*incomingOp, sendConn)
			}
		}
	}
}

func (e *Editor) adjustCursorAfterRemoteOp(oldPosition int, op *crdt.Operation) {
	switch op.Type {
	case crdt.Insert:
		if op.Position <= oldPosition {
			e.LocalCursor.Position++
		}
	case crdt.Delete:
		if op.Position < oldPosition {
			e.LocalCursor.Position--
		} else if op.Position == oldPosition {
			e.LocalCursor.Position = e.findValidCursorPosition(oldPosition)
		}
	}
	e.updateLocalCursor()
}

func (e *Editor) findValidCursorPosition(position int) int {
	content := []rune(e.RGA.GetText())
	if position >= len(content) {
		return len(content)
	}
	for position > 0 && content[position] == rune(0) {
		position--
	}
	return position
}

func (e *Editor) updateRemoteCursor(id string, position int) {
	e.remoteCursorMu.Lock()
	defer e.remoteCursorMu.Unlock()
	cursor, exists := e.RemoteCursors[id]
	if !exists {
		e.guestCounter++
		cursor = CursorInfo{
			Username:   fmt.Sprintf("guest%d", e.guestCounter),
			ThemeIndex: e.nextThemeIndex,
		}
		e.nextThemeIndex = (e.nextThemeIndex + 1) % len(e.Theme.UserThemes)
	}
	cursor.Position = position
	cursor.LastMove = time.Now()
	e.RemoteCursors[id] = cursor
}

func (e *Editor) renderCursorWithName(cursor CursorInfo) string {
	return e.Theme.RenderCursor(e.IsSharedSession, cursor.ThemeIndex)
}

func (e *Editor) HandleConnections() {
	for {
		newConn := <-e.NewConnection
		e.IsSharedSession = true
		e.Update <- struct{}{}
		go e.reciveInput(newConn)
	}
}

func (e *Editor) GetLineNumbers() string {
	var lineNumbers strings.Builder
	lineNumber := 1
	lineNumbers.WriteString(fmt.Sprintf("%d\n", lineNumber))

	for _, char := range e.RGA.GetText() {
		if char == '\n' {
			lineNumber++
			lineNumbers.WriteString(fmt.Sprintf("%d\n", lineNumber))
		}
	}

	return lineNumbers.String()
}

func (e *Editor) RenderContent() string {
	crdt.InsertM.Lock()
	defer crdt.InsertM.Unlock()

	var output strings.Builder
	content := e.RGA.GetText()

	lines := strings.Split(content, "\n")
	lineNumberWidth := len(fmt.Sprintf("%d", len(lines)))
	totalLines := e.Viewport.Height - 2 // Subtracting 2 for header and footer

	for i := 0; i < totalLines; i++ {
		lineNumber := ""
		line := ""

		if i < len(lines) {
			lineNumber = fmt.Sprintf("%*d", lineNumberWidth, i+1)
			line = lines[i]
		} else {
			lineNumber = strings.Repeat(" ", lineNumberWidth-1) + "~"
		}

		renderedLineNumber := e.Theme.RenderLineNumber(lineNumber, lineNumberWidth)

		if i < len(lines) {
			renderedLine := e.renderLineWithCursors(line, i)

			output.WriteString(renderedLineNumber + renderedLine + "\n")
		} else {
			output.WriteString(renderedLineNumber + "\n")
		}
	}

	headerMsg := fmt.Sprintf("File: %s", e.FilePath)
	if e.Network.IsHost {
		headerMsg += fmt.Sprintf(" Clients: %d", len(e.Network.Clients))
	}
	if e.Network.Host != nil {
		headerMsg += fmt.Sprintf(" Session: %s", e.Network.CurrentSession)
	}

	header := e.Theme.RenderHeader(headerMsg)
	footer := e.Theme.RenderStatusBar(e.Error)

	e.Error = ""

	content = fmt.Sprintf("%s\n%s%s", header, output.String(), footer)

	return lipgloss.NewStyle().MaxWidth(e.Viewport.Width).MaxHeight(e.Viewport.Height).Render(content)
}

func (e *Editor) RenderDocumentWithoutLineNumbers() string {
	return e.RGA.GetTextWithOutTomestone()
}

func (e *Editor) renderLineWithCursors(line string, lineIndex int) string {
	var result strings.Builder
	lineStartIndex := e.getLineStartIndex(lineIndex)

	c_cursior := e.RGA.ConvertCursior(e.LocalCursor.Position)

	for colIndex, ch := range line {
		absoluteIndex := lineStartIndex + colIndex

		if absoluteIndex == c_cursior {
			result.WriteString(e.renderCursorWithName(e.LocalCursor))
		}

		e.remoteCursorMu.RLock()
		for _, remoteCursor := range e.RemoteCursors {
			c_remoteCursior := e.RGA.ConvertCursior(remoteCursor.Position)
			if absoluteIndex == c_remoteCursior {
				result.WriteString(e.renderCursorWithName(remoteCursor))
			}
		}
		e.remoteCursorMu.RUnlock()

		result.WriteRune(ch)
	}

	// Check for cursors at the end of the line
	if lineStartIndex+len(line) == c_cursior {
		result.WriteString(e.renderCursorWithName(e.LocalCursor))
	}

	e.remoteCursorMu.RLock()
	for _, remoteCursor := range e.RemoteCursors {
		c_remoteCursior := e.RGA.ConvertCursior(remoteCursor.Position)
		if lineStartIndex+len(line) == c_remoteCursior {
			result.WriteString(e.renderCursorWithName(remoteCursor))
		}
	}
	e.remoteCursorMu.RUnlock()

	colorText := e.SyntaxDef.EmiteColorText(line, result.String())

	return colorText
}

func (e *Editor) getLineStartIndex(lineIndex int) int {
	content := e.RGA.GetText()
	lines := strings.Split(content, "\n")
	startIndex := 0
	for i := 0; i < lineIndex; i++ {
		startIndex += len(lines[i]) + 1 // +1 for the newline character
	}
	return startIndex
}

func (e *Editor) getCursorLineAndColumn() (int, int) {
	content := e.RGA.GetText()
	line := 1
	col := 1

	for i, ch := range content {
		if i == e.LocalCursor.Position {
			break
		}
		if ch == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}
