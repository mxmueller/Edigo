package crdt

import (
	"strings"
)

// CRDTCharacter represents a character in the CRDT model
type CRDTCharacter struct {
	ID      string
	Value   string
	Visible bool
	PrevID  string
	NextID  string
}

// CRDT represents the entire text document using a CRDT model
type CRDT struct {
	Chars map[string]*CRDTCharacter
	Head  string
}

// NewCRDT initializes a new CRDT instance
func NewCRDT() *CRDT {
	return &CRDT{
		Chars: make(map[string]*CRDTCharacter),
		Head:  "",
	}
}

// Insert inserts a new character into the CRDT
func (c *CRDT) Insert(id, value, prevID, nextID string) {
	char := &CRDTCharacter{
		ID:      id,
		Value:   value,
		Visible: true,
		PrevID:  prevID,
		NextID:  nextID,
	}
	c.Chars[id] = char

	if prevID == "" {
		// Insert at the beginning
		char.NextID = c.Head
		c.Head = id
	} else {
		// Insert in the middle or end
		prevChar := c.Chars[prevID]
		nextChar := c.Chars[prevChar.NextID]

		prevChar.NextID = id
		if nextChar != nil {
			char.NextID = nextChar.ID
		}
	}
}

// Delete marks a character as invisible
func (c *CRDT) Delete(id string) {
	if char, exists := c.Chars[id]; exists {
		char.Visible = false
	}
}

// String returns the current state of the CRDT as a string
func (c *CRDT) String() string {
	var builder strings.Builder
	currentID := c.Head
	for currentID != "" {
		char := c.Chars[currentID]
		if char.Visible {
			builder.WriteString(char.Value)
		}
		currentID = char.NextID
	}
	return builder.String()
}
