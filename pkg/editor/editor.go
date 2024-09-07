package editor

import (
	"bytes"
	"edigo/pkg/crdt"
	"edigo/pkg/network"
	"edigo/pkg/theme"
	"encoding/gob"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type RemoteChange struct{}

type CursorInfo struct {
	Position   int
	LastMove   time.Time
	ShowName   bool
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
	LocalCursor     CursorInfo
	RemoteCursors   map[string]CursorInfo
	remoteCursorMu  sync.RWMutex
	updateTicker    *time.Ticker
	nextThemeIndex  int
	IsSharedSession bool
}

func NewEditor(content string, siteID string, theme *theme.Theme) *Editor {
	rga := crdt.NewRGA(siteID)
	for _, char := range content {
		rga.LocalInsert(char)
	}

	newConnection := make(chan net.Conn, 1)
	network := network.NewNetwork()
	network.NewConnection = newConnection

	editor := &Editor{
		RGA:           rga,
		Network:       network,
		NewConnection: newConnection,
		Theme:         theme,
		Viewport:      viewport.New(80, 24),
		LocalCursor: CursorInfo{
			Position:   0,
			LastMove:   time.Now(),
			ShowName:   false,
			Username:   "You",
			ThemeIndex: 0,
		},
		RemoteCursors:   make(map[string]CursorInfo),
		updateTicker:    time.NewTicker(100 * time.Millisecond),
		Update:          make(chan struct{}, 1),
		nextThemeIndex:  1,
		IsSharedSession: false,
	}

	go editor.handleTicker()

	return editor
}

func (e *Editor) handleTicker() {
	for range e.updateTicker.C {
		updated := false
		if !e.LocalCursor.ShowName && time.Since(e.LocalCursor.LastMove) > 1500*time.Millisecond {
			e.LocalCursor.ShowName = true
			updated = true
		}
		e.remoteCursorMu.Lock()
		for id, cursor := range e.RemoteCursors {
			if !cursor.ShowName && time.Since(cursor.LastMove) > 1500*time.Millisecond {
				cursor.ShowName = true
				e.RemoteCursors[id] = cursor
				updated = true
			}
		}
		e.remoteCursorMu.Unlock()
		if updated {
			e.Update <- struct{}{}
		}
	}
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
	e.LocalCursor.ShowName = false
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
		cursor = CursorInfo{
			Username:   fmt.Sprintf("User-%s", id[:5]),
			ThemeIndex: e.nextThemeIndex,
		}
		e.nextThemeIndex = (e.nextThemeIndex + 1) % len(e.Theme.UserThemes)
	}
	cursor.Position = position
	cursor.LastMove = time.Now()
	cursor.ShowName = false
	e.RemoteCursors[id] = cursor
}

func (e *Editor) renderCursorWithName(cursor CursorInfo) string {
	cursorRender := e.Theme.RenderCursor(e.IsSharedSession, cursor.ThemeIndex)
	if e.IsSharedSession && cursor.ShowName {
		return cursorRender + e.Theme.RenderUsername(cursor.Username, cursor.ThemeIndex)
	}
	return cursorRender
}

func (e *Editor) HandleConnections() {
	for {
		newConn := <-e.NewConnection
		e.IsSharedSession = true
		e.Update <- struct{}{}
		go e.reciveInput(newConn)
	}
}

func (e *Editor) RenderDocument() string {
	crdt.InsertM.Lock()
	defer crdt.InsertM.Unlock()

	var result strings.Builder
	content := e.RGA.GetText()

	for i, ch := range content {
		if i == e.LocalCursor.Position {
			result.WriteString(e.renderCursorWithName(e.LocalCursor))
		}
		e.remoteCursorMu.RLock()
		for _, remoteCursor := range e.RemoteCursors {
			if i == remoteCursor.Position {
				result.WriteString(e.renderCursorWithName(remoteCursor))
			}
		}
		e.remoteCursorMu.RUnlock()
		result.WriteRune(ch)
	}

	if e.LocalCursor.Position == len(content) {
		result.WriteString(e.renderCursorWithName(e.LocalCursor))
	}
	e.remoteCursorMu.RLock()
	for _, remoteCursor := range e.RemoteCursors {
		if remoteCursor.Position == len(content) {
			result.WriteString(e.renderCursorWithName(remoteCursor))
		}
	}
	e.remoteCursorMu.RUnlock()

	return result.String()
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

func (e *Editor) RenderDocumentWithoutLineNumbers() string {
	return e.RGA.GetText()
}

func (e *Editor) RenderContent() string {
	lineNumbers := e.GetLineNumbers()
	document := e.RenderDocument()

	lineNumberWidth := len(fmt.Sprintf("%d", strings.Count(document, "\n")+1))

	var output strings.Builder

	lines := strings.Split(document, "\n")
	numberLines := strings.Split(lineNumbers, "\n")

	totalLines := e.Viewport.Height - 2 // Subtracting 2 for header and footer

	for i := 0; i < totalLines; i++ {
		if i < len(lines) {
			lineNumber := ""
			if i < len(numberLines) {
				lineNumber = e.Theme.RenderLineNumber(numberLines[i], lineNumberWidth)
			} else {
				lineNumber = e.Theme.RenderLineNumber("", lineNumberWidth)
			}

			line := e.Theme.BaseStyle.Render(lines[i])
			output.WriteString(lineNumber + line + "\n")
		} else {
			output.WriteString(e.Theme.RenderLineNumber("~", lineNumberWidth) + "\n")
		}
	}

	headerMsg := ""
	if e.Network.IsHost {
		headerMsg = fmt.Sprintf("File: %s Clients: %d", e.FilePath, len(e.Network.Clients))
	} else if e.Network.CurrentSession == "" {
		headerMsg = fmt.Sprintf("File: %s", e.FilePath)
	}
	if e.Network.Host != nil {
		headerMsg = fmt.Sprintf("Session: %s", e.Network.CurrentSession)
	}

	header := e.Theme.RenderHeader(headerMsg)
	footer := e.Theme.RenderStatusBar(e.Error)

	e.Error = ""

	content := fmt.Sprintf("%s\n%s%s", header, output.String(), footer)

	return lipgloss.NewStyle().MaxWidth(e.Viewport.Width).MaxHeight(e.Viewport.Height).Render(content)
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
