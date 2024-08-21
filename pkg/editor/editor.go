package editor

import (
	"bytes"
	"edigo/pkg/crdt"
	"edigo/pkg/network"
	"encoding/gob"
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type RemoteChange struct{}

type Editor struct {
	RGA *crdt.RGA
    Viewport *viewport.Model
    Network *network.Network
    FilePath string
    Update chan struct{} // UpdateIU
    NewConnection chan net.Conn
    Error string
}

func NewEditor(content string, siteID string) *Editor {
	rga := crdt.NewRGA(siteID)
	for _, char := range content {
		rga.LocalInsert(char)
	}

    newConnection := make(chan net.Conn, 1)
    network := network.NewNetwork()
    network.NewConnection = newConnection

	return &Editor{
		RGA: rga,
        Network: network,
        NewConnection: newConnection,
	}
}

func (e *Editor) InsertCharacter(ch rune) {
    op := e.RGA.LocalInsert(ch)
    e.sendToRemote(op)
}

func (e *Editor) DeleteCharacterBeforeCursor() {
    op := e.RGA.LocalDelete()
    e.sendToRemote(op)
}

func (e *Editor) MoveCursorLeft() {
	e.RGA.MoveCursorLeft()
}

func (e *Editor) MoveCursorRight() {
	e.RGA.MoveCursorRight()
}

func (e *Editor) MoveCursorUp() {
	e.RGA.MoveCursorUp()
}

func (e *Editor) MoveCursorDown() {
	e.RGA.MoveCursorDown()
}

func (e *Editor) sendToRemote(op crdt.Operation){

    if e.Network.IsHost{
        for _, conn := range e.Network.Clients{
            e.Network.SendOperation(op, conn)
        }
    }else if e.Network.Host != nil{
            e.Network.SendOperation(op, e.Network.Host)
    }
}

func (e *Editor) HandleConnections() {

    for {
        newConn := <- e.NewConnection
        go e.reciveInput(newConn)
    }
}

func (e *Editor) reciveInput(conn net.Conn) {

        defer conn.Close()
            // Empfange Nachrichten und gib sie aus
        for {
            buf := make([]byte, 2064)
            _, err := conn.Read(buf)
            if err != nil { // TODO make error msg check
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
	var result strings.Builder
	content := e.RGA.GetText()

	for i, ch := range content {
		if i == e.RGA.CursorPosition {
			result.WriteRune('█') // Cursor
		}
		result.WriteRune(ch)
	}

	if e.RGA.CursorPosition == len(content) {
		result.WriteRune('█') // Cursor End
	}

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

	lineNumberWidth := len(fmt.Sprintf("%d", strings.Count(document, "\n")+1)) + 2

	var output strings.Builder

	lines := strings.Split(document, "\n")
	numberLines := strings.Split(lineNumbers, "\n")

	totalLines := e.Viewport.Height - 2 // Subtracting 2 for filename and empty line

	for i := 0; i < totalLines; i++ {
		if i < len(lines) {
			if i < len(numberLines) {
				output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, numberLines[i]))
			} else {
				output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, ""))
			}
			output.WriteString(lines[i])
		} else {
			output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, "~"))
		}
		output.WriteString("\n")
	}

    headerMsg := "" 

    if(e.Network.IsHost || e.Network.CurrentSession == ""){
        headerMsg = fmt.Sprintf("File: %s", e.FilePath)
    }
    if(e.Network.Host != nil){
        headerMsg = fmt.Sprintf("Session: %s", e.Network.CurrentSession)
    }


	header := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1).
		Render(headerMsg)

	fotter := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
        BorderTop(true).
		Padding(0, 1).
		Render(e.Error)

    e.Error = ""

	return fmt.Sprintf("%s\n%s\n%s", header, output.String(), fotter)
}
