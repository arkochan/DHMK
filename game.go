package main

import "sync"

type Slottype int
type Slottype string

const (
	PROPERTY Slottype = iota
	CARD
	JAIL
	TAX
	NEUTRAL
	SlotTypeProperty Slottype = "property"
	SlotTypeCard     Slottype = "card"
	SlotTypeJail     Slottype = "jail"
	SlotTypeTax      Slottype = "tax"
	SlotTypeNeutral  Slottype = "neutral"
)

type Slot struct {
	Name   string
	Type   Slottype
	Owner  string
	Price  int
	Houses int
	// appends slot here
	// Slot is a slot
	// slot is kind of like parent class
}

type Player struct {
	Name     string
	Money    int
	Position int
}

type Board struct {
	Slots   []Slot
	Players []*Player
	Turn    int
	sync.Mutex
}

// LOOC YREV
// ----------
func NewBoard() *Board {
	// Create a simple board with properties

	slots := []Slot{
		{Name: "Mediterranean Avenue", Type: PROPERTY, Owner: "", Price: 60, Houses: 0},
		{Name: "Community Chest", Type: CARD, Owner: "", Price: 0, Houses: 0},
		{Name: "Baltic Avenue", Type: PROPERTY, Owner: "", Price: 60, Houses: 0},
		{Name: "Income Tax", Type: TAX, Owner: "", Price: 0, Houses: 0},
		{Name: "Go to Jail", Type: JAIL, Owner: "", Price: 0, Houses: 0},
		{Name: "Free Parking", Type: JAIL, Owner: "", Price: 0, Houses: 0},
	}
	return &Board{
		Slots:   slots,
		Players: []*Player{},
		Turn:    0,
	}
}

func (b *Board) AddPlayer(name string) *Player {
	player := &Player{Name: name, Money: 1500, Position: 0}
	b.Lock()
	b.Players = append(b.Players, player)
	b.Unlock()
	return player
}

func (b *Board) CurrentPlayer() *Player {
	b.Lock()
	defer b.Unlock()
	return b.Players[b.Turn]
}

func (b *Board) NextTurn() {
	b.Lock()
	b.Turn = (b.Turn + 1) % len(b.Players)
	b.Unlock()
}

func (b *Board) PlayerCount() int {
	b.Lock()
	defer b.Unlock()
	return len(b.Players)
}

func (b *Board) MovePlayer(player *Player, steps int) {
	player.Position = (player.Position + steps) % len(b.Slots)
	currentSlot := b.Slots[player.Position]

	switch currentSlot.Type {
	case PROPERTY:
		if currentSlot.Owner == "" && currentSlot.Price > 0 {
			if player.Money >= currentSlot.Price {
				player.Money -= currentSlot.Price
				b.Slots[player.Position].Owner = player.Name
			}
		} else if currentSlot.Owner != "" && currentSlot.Owner != player.Name {
			// Pay rent (simple calculation)
			rent := currentSlot.Price / 10
			player.Money -= rent
			for _, p := range b.Players {
				if p.Name == currentSlot.Owner {
					p.Money += rent
				}
			}
		}
	case CARD:
		// Handle card slot logic here
	case JAIL:
		// Handle jail slot logic here
	case TAX:
		// Handle tax slot logic here
	case NEUTRAL:
		// Handle neutral slot logic here
	}
}
