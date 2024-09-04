package crdt

import (
	"fmt"
	"math/rand" // Import für die Zufallszahlengenerierung
	"strings"
	"sync"
	"time"
)

type OperationType int

const (
	Insert OperationType = iota
	Delete
    Move
)

var (
	InsertM sync.Mutex
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
    RemoteCursors map[string]int
}

func NewRGA(site string) *RGA {
	return &RGA{
		Elements:       []Element{},
		Site:           site,
		Clock:          0,
		CursorPosition: 0,
        RemoteCursors: make(map[string]int),
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
    rga.MoveCursorRight()

	return op
}

func (rga *RGA) LocalDelete() Operation {
	if rga.CursorPosition > 0 {
        rga.MoveCursorLeft()
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

func (rga *RGA) SetRemoteCursor(op Operation) {
    rga.RemoteCursors[op.ID] = op.Position
}

func (rga *RGA) ApplyOperation(op Operation) {
    InsertM.Lock()
	switch op.Type {
	case Insert:
		rga.RemoteInsert(op)
	case Delete:
		rga.RemoteDelete(op)
	case Move:
        rga.SetRemoteCursor(op)
	}
    InsertM.Unlock()
}

func (rga *RGA) GetText() string {
	var result strings.Builder
	for _, elem := range rga.Elements {
		if !elem.Tombstone {
			result.WriteRune(elem.Character)
		} else{
			result.WriteRune(rune(0))
        }
	}
	return result.String()
}

func (rga *RGA) MoveCursorLeft() Operation{
	for rga.CursorPosition > 0 {
		rga.CursorPosition--
		if !rga.Elements[rga.CursorPosition].Tombstone {
			break
		}
	}
    return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}
}

func (rga *RGA) MoveCursorRight() Operation{
	for rga.CursorPosition < len(rga.Elements) {
		rga.CursorPosition++
        if rga.CursorPosition >= len(rga.Elements) {
            return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}
        }
		if !rga.Elements[rga.CursorPosition].Tombstone {
			break
		}
	}
    return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}
}

func (rga *RGA) MoveCursorUp() Operation {
    if rga.CursorPosition == 0 {
		return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}

	}

	for rga.CursorPosition > 0 {
        rga.MoveCursorLeft()
		if rga.Elements[rga.CursorPosition].Character == '\n' {
			break
		}
	}

	for rga.CursorPosition > 0 {
        rga.MoveCursorLeft()
		if rga.Elements[rga.CursorPosition].Character == '\n' {
            rga.MoveCursorRight()
			break
		}
	}
	return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}
}

// map cursor
func (rga *RGA) MoveCursorDown() Operation {
	for rga.CursorPosition < len(rga.Elements) && rga.Elements[rga.CursorPosition].Character != '\n' {
            rga.MoveCursorRight()
	}

	if rga.CursorPosition < len(rga.Elements) {
            rga.MoveCursorRight()
	}

	for rga.CursorPosition < len(rga.Elements) && rga.Elements[rga.CursorPosition].Character != '\n' {
            rga.MoveCursorRight()
	}
	return Operation{Type: Move, ID: rga.Site, Character: 0, Position: rga.CursorPosition}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
