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
	"time"

	"github.com/charmbracelet/bubbles/viewport"
)

type RemoteChange struct{}

type Editor struct {
	RGA           *crdt.RGA
	Viewport      *viewport.Model
	Network       *network.Network
	FilePath      string
	Update        chan struct{}
	NewConnection chan net.Conn
	Error         string
	Theme         *theme.Theme
}

func NewEditor(content string, siteID string, theme *theme.Theme) *Editor {
	rga := crdt.NewRGA(siteID)
	for _, char := range content {
		rga.LocalInsert(char)
	}

	newConnection := make(chan net.Conn, 1)
	network := network.NewNetwork()
	network.NewConnection = newConnection

	return &Editor{
		RGA:           rga,
		Network:       network,
		NewConnection: newConnection,
		Theme:         theme,
	}
}

func (e *Editor) InsertCharacter(ch rune) {
    op := e.RGA.LocalInsert(ch)
    e.sendToRemote(op)
    go e.sendToRemote(crdt.Operation{Type: crdt.Move, ID: e.RGA.Site, Character: 0, Position: e.RGA.CursorPosition})
}

func (e *Editor) DeleteCharacterBeforeCursor() {
    op := e.RGA.LocalDelete()
    e.sendToRemote(op)
    go e.sendToRemote(crdt.Operation{Type: crdt.Move, ID: e.RGA.Site, Character: 0, Position: e.RGA.CursorPosition})
}

func (e *Editor) MoveCursorLeft() {
    op := e.RGA.MoveCursorLeft()
    e.sendToRemote(op)
}

func (e *Editor) MoveCursorRight() {
    op := e.RGA.MoveCursorRight()
    e.sendToRemote(op)
}

func (e *Editor) MoveCursorUp() {
	op := e.RGA.MoveCursorUp()
    e.sendToRemote(op)
}

func (e *Editor) MoveCursorDown() {
    op := e.RGA.MoveCursorDown()
    e.sendToRemote(op)
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

func (e *Editor) HandleConnections() {
	for {
		newConn := <-e.NewConnection
		e.Update <- struct{}{}
		go e.reciveInput(newConn)
	}
}

func (e *Editor) reciveInput(conn net.Conn) {
    defer conn.Close()
            // Empfange Nachrichten und gib sie aus
        for {
            buf := make([]byte, 256)

            network.ReadM.Lock()
            conn.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
            _, err := conn.Read(buf)
            network.ReadM.Unlock()

            if err != nil { // TODO make error msg check

                if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                    continue
                }

                if e.Network.IsHost{
                    // client hat Verbindung geschlossen
                    e.Network.RemoveClient(conn)
                }
                if e.Network.Host != nil {
                    e.Network.HostClosedSession()
                // falls man will sollte man hier den alten Content wieder reinladen
            }
                e.Update<- struct{}{}
                return
            }
            tmpbuff := bytes.NewBuffer(buf)
            incomingOp := new(crdt.Operation)

            gobobj := gob.NewDecoder(tmpbuff)
            gobobj.Decode(incomingOp)
            e.RGA.ApplyOperation(*incomingOp)
            e.Update <- struct{}{}

            if e.Network.IsHost{ // forward to other editor
                for _, sendConn := range e.Network.Clients{
                    if sendConn == conn{
                        continue
                    } 
                    e.Network.SendOperation(*incomingOp, sendConn)
                }
        }
    }
}


func (e *Editor) RenderDocument() string {
    crdt.InsertM.Lock()
	var result strings.Builder
	content := e.RGA.GetText()

	for i, ch := range content {
		if i == e.RGA.CursorPosition {
		    result.WriteString(e.Theme.RenderCursor(" "))
		}
        for _, j := range e.RGA.RemoteCursors{
		    if i == j {
		    result.WriteString(e.Theme.RenderCursor(" "))
            }
        }
		result.WriteRune(ch)
	}

	if e.RGA.CursorPosition == len(content) {
		result.WriteString(e.Theme.RenderCursor(" "))
	}
    for _, j := range e.RGA.RemoteCursors{
        if len(content) == j {
		result.WriteString(e.Theme.RenderCursor(" "))
        }
    }

    crdt.InsertM.Unlock()
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

	totalLines := e.Viewport.Height - 2 // Subtracting 2 for filename and empty line

	for i := 0; i < totalLines; i++ {
		if i < len(lines) {
			if i < len(numberLines) {
				output.WriteString(e.Theme.RenderLineNumber(numberLines[i], lineNumberWidth))
			} else {
				output.WriteString(e.Theme.RenderLineNumber("", lineNumberWidth))
			}
			output.WriteString(e.Theme.BaseStyle.Render(lines[i]))
		} else {
			output.WriteString(e.Theme.RenderLineNumber("~", lineNumberWidth))
		}
		output.WriteString("\n")
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

	return fmt.Sprintf("%s\n%s\n%s", header, output.String(), footer)
}
