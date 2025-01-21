package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type Slottype string

const (
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
	Name       string
	Money      int
	Position   int
	MoveLocked bool
}

type Board struct {
	Slots   []Slot
	Players []*Player
	Turn    int
	sync.Mutex
}

type GameTradeBody struct {
	Give TradeDetails `json:"give,omitempty"`
	Take TradeDetails `json:"take,omitempty"`
}

type TradeDetails struct {
	Property []string `json:"property,omitempty"`
	Money    int      `json:"money,omitempty"`
	Cards    []string `json:"cards,omitempty"`
}

// LOOC YREV
// ----------
func NewBoard() *Board {
	// Create a simple board with properties

	slots := []Slot{
		{Name: "Mediterranean Avenue", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		{Name: "Arkochan Avenue", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		{Name: "Chittagong", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		// {Name: "Community Chest", Type: SlotTypeCard, Owner: "", Price: 0, Houses: 0},
		{Name: "Hell Yeah Avenue", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		{Name: "Nicsu York", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		{Name: "MiniSoda", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		{Name: "Ohio", Type: SlotTypeProperty, Owner: "", Price: 60, Houses: 0},
		// {Name: "Income Tax", Type: SlotTypeTax, Owner: "", Price: 0, Houses: 0},
		// {Name: "Go to Jail", Type: SlotTypeJail, Owner: "", Price: 0, Houses: 0},
		// {Name: "Free Parking", Type: SlotTypeNeutral, Owner: "", Price: 0, Houses: 0},
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

func (b *Board) RollDice() int {
	return rand.Intn(12) + 1
}

func (b *Board) HandleAction(player *Player, message Message) (string, string, error) {
	// if players turn
	if player.Name == b.CurrentPlayer().Name {
		switch message.Action {
		case ActionGo:
			return b.RollPlayer(player)
		case ActionBuy:
			return b.BuyProperty(player)
		case ActionEndTurn:
			return b.EndTurn(player)

			// TODO:
			// case ActionMortgage:
			// case ActionUseCard:
			// case ActionBuyHouse:
		default:
			return "", "", fmt.Errorf("invalid action: %s", message.Action)
		}
	} else {
		// if not players turn

		switch message.Action {
		case ActionTrade:

		case ActionForfeitGame:

		default:
			return "", "", fmt.Errorf("invalid action: %s", message.Action)

		}
	}
	return "", "", nil
}

func (b *Board) BuyProperty(player *Player) (string, string, error) {
	slot := b.Slots[player.Position]
	if slot.Owner != "" {
		return "", "", fmt.Errorf("slot already owned")
	}
	if player.Money < slot.Price {
		return "", "", fmt.Errorf("insufficient funds")
	}
	// TODO:
	// Create a transaction function
	player.Money -= slot.Price
	slot.Owner = player.Name
	return fmt.Sprintf("%s bought %s for %d", player.Name, slot.Name, slot.Price), "", nil
}

func (b *Board) MovePlayer(player *Player, steps int) (string, string, error) {
	player.Position = (player.Position + steps) % len(b.Slots)
	currentSlot := b.Slots[player.Position]

	switch currentSlot.Type {
	case SlotTypeProperty:
		if currentSlot.Owner == "" && currentSlot.Price > 0 {
			// prompt user
			return "", fmt.Sprintf("Want to buy %s for %d?", currentSlot.Name, currentSlot.Price), nil
		} else if currentSlot.Owner != "" && currentSlot.Owner != player.Name {
			// Pay rent (simple calculation)
			rent := currentSlot.Price / 10
			player.Money -= rent
			for _, p := range b.Players {
				if p.Name == currentSlot.Owner {
					p.Money += rent
				}
			}
			return fmt.Sprintf("%s paid %d rent to %s", player.Name, rent, currentSlot.Owner), "", nil
		}
	case SlotTypeCard:
		// Handle card slot logic here
		return "", "", fmt.Errorf("%s", currentSlot.Type)
	case SlotTypeJail:
		// Handle jail slot logic here
	case SlotTypeTax:
		// Handle tax slot logic here
	case SlotTypeNeutral:
		// Handle neutral slot logic here
	}
	return "", "", fmt.Errorf("invalid slot type: %s", currentSlot.Type)
}

func (b *Board) RollPlayer(player *Player) (string, string, error) {
	steps := b.RollDice()
	// TODO:
	// Handle Double
	if player.MoveLocked {
		return "", "", fmt.Errorf("cant move now")
	}
	// TODO:
	return b.MovePlayer(player, steps)
}

