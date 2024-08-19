package crdt

import (
	"fmt"
	"math/rand" // Import für die Zufallszahlengenerierung
	"strings"
	"time"
)

type OperationType int

const (
	Insert OperationType = iota
	Delete
)

type Element struct {
	ID        string
	Character rune
	Tombstone bool
}

type Operation struct {
	Type      OperationType
	ID        string
	Character rune
	Position  int
}

type RGA struct {
	Elements       []Element
	Site           string
	Clock          int
	CursorPosition int
}

func NewRGA(site string) *RGA {
	return &RGA{
		Elements:       []Element{},
		Site:           site,
		Clock:          0,
		CursorPosition: 0,
	}
}

func (rga *RGA) generateID() string {
	rga.Clock++
	return fmt.Sprintf("%s-%d-%d", rga.Site, rga.Clock, rand.Intn(1000))
}

func (rga *RGA) LocalInsert(char rune) Operation {
	id := rga.generateID()
	newElement := Element{ID: id, Character: char, Tombstone: false}

	if rga.CursorPosition >= len(rga.Elements) {
		rga.Elements = append(rga.Elements, newElement)
	} else {
		rga.Elements = append(rga.Elements[:rga.CursorPosition+1], rga.Elements[rga.CursorPosition:]...)
		rga.Elements[rga.CursorPosition] = newElement
	}

	op := Operation{Type: Insert, ID: id, Character: char, Position: rga.CursorPosition}
	rga.CursorPosition++

	return op
}

func (rga *RGA) LocalDelete() Operation {
	if rga.CursorPosition > 0 {
		rga.CursorPosition--
		rga.Elements[rga.CursorPosition].Tombstone = true
		op := Operation{Type: Delete, ID: rga.Elements[rga.CursorPosition].ID, Position: rga.CursorPosition}

		return op
	}
	return Operation{}
}

func (rga *RGA) RemoteInsert(op Operation) {
	newElement := Element{ID: op.ID, Character: op.Character, Tombstone: false}

	if op.Position >= len(rga.Elements) {
		rga.Elements = append(rga.Elements, newElement)
	} else {
		rga.Elements = append(rga.Elements[:op.Position+1], rga.Elements[op.Position:]...)
		rga.Elements[op.Position] = newElement
	}
}

func (rga *RGA) RemoteDelete(op Operation) {
	for i, elem := range rga.Elements {
		if elem.ID == op.ID {
			rga.Elements[i].Tombstone = true
			break
		}
	}
}

func (rga *RGA) ApplyOperation(op Operation) {
	switch op.Type {
	case Insert:
		rga.RemoteInsert(op)
	case Delete:
		rga.RemoteDelete(op)
	}
}

func (rga *RGA) GetText() string {
	var result strings.Builder
	for _, elem := range rga.Elements {
		if !elem.Tombstone {
			result.WriteRune(elem.Character)
		}
	}
	return result.String()
}

func (rga *RGA) MoveCursorLeft() {
	if rga.CursorPosition > 0 {
		rga.CursorPosition--
	}
}

func (rga *RGA) MoveCursorRight() {
	if rga.CursorPosition < len(rga.Elements) {
		rga.CursorPosition++
	}
}

func (rga *RGA) MoveCursorUp() {
	if rga.CursorPosition == 0 {
		return
	}

	for rga.CursorPosition > 0 {
		rga.CursorPosition--
		if rga.Elements[rga.CursorPosition].Character == '\n' {
			break
		}
	}

	for rga.CursorPosition > 0 {
		rga.CursorPosition--
		if rga.Elements[rga.CursorPosition].Character == '\n' {
			rga.CursorPosition++
			break
		}
	}
}

// map cursor
func (rga *RGA) MoveCursorDown() {
	for rga.CursorPosition < len(rga.Elements) && rga.Elements[rga.CursorPosition].Character != '\n' {
		rga.CursorPosition++
	}

	if rga.CursorPosition < len(rga.Elements) {
		rga.CursorPosition++
	}

	for rga.CursorPosition < len(rga.Elements) && rga.Elements[rga.CursorPosition].Character != '\n' {
		rga.CursorPosition++
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
