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
	Owner  IdType
	Price  int
	Houses int
	// appends slot here
	// Slot is a slot
	// slot is kind of like parent class
}

type Player struct {
	Id         IdType
	Name       string
	Money      int
	Position   int
	MoveLocked bool
}

type Board struct {
	Slots   []Slot
	Players []*Player
	Trades  []*GameTradeBody
	Turn    int
	sync.Mutex
	MoveLocked bool
}

type IdType *int

type (
	GameTradeBody struct {
		To     IdType       `json:"to" validate:"required"`
		Id     IdType       `json:"id" validate:"required"` // `required` ensures this field must be present
		From   IdType       `json:"from" validate:"required"`
		Give   TradeDetails `json:"give,omitempty"`
		Take   TradeDetails `json:"take,omitempty"`
		Accept bool         `json:"accept,omitempty"`
		Active bool         `json:"active,omitempty"`
	}
)

type TradeDetails struct {
	Property []IdType `json:"property,omitempty"`
	Money    int      `json:"money,omitempty"`
	Cards    []IdType `json:"cards,omitempty"`
}

// LOOC YREV
// ----------
func NewBoard() *Board {
	// Create a simple board with properties

	slots := []Slot{
		{Name: "Mediterranean Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		{Name: "Arkochan Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		{Name: "Chittagong", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		// {Name: "Community Chest", Type: SlotTypeCard, Owner: nil, Price: 0, Houses: 0},
		{Name: "Hell Yeah Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		{Name: "Nicsu York", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		{Name: "MiniSoda", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		{Name: "Ohio", Type: SlotTypeProperty, Owner: nil, Price: 60, Houses: 0},
		// {Name: "Income Tax", Type: SlotTypeTax, Owner: nil, Price: 0, Houses: 0},
		// {Name: "Go to Jail", Type: SlotTypeJail, Owner: nil, Price: 0, Houses: 0},
		// {Name: "Free Parking", Type: SlotTypeNeutral, Owner: nil, Price: 0, Houses: 0},
	}
	return &Board{
		Slots:   slots,
		Players: []*Player{},
		Turn:    0,
	}
}

func (b *Board) AddPlayer(name string) *Player {
	newId := len(b.Players)
	player := &Player{Name: name, Money: 1500, Position: 0, Id: &newId}
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

func (b *Board) PlayerNames() []string {
	b.Lock()
	defer b.Unlock()
	names := []string{}
	for _, player := range b.Players {
		names = append(names, player.Name)
	}
	return names
}

type TransactionType int

const (
	TransactionBankToPlayer TransactionType = iota
	TransactionPlayerToBank
	TransactionPlayerToPlayer
)

// a transaction function sender reciver amount
// sender/ reciver can be -> player / bank / all players
func (b *Board) TransferBankToPlayer(receiver *Player, amount int) {
	receiver.Money += amount
}

func (b *Board) TransferPlayerToBank(sender *Player, amount int) error {
	if sender.Money < amount {
		return fmt.Errorf("insufficient funds")
	}
	sender.Money -= amount
	return nil
}

func (b *Board) TransferPlayerToPlayer(sender *Player, receiver *Player, amount int) error {
	if sender.Money < amount {
		return fmt.Errorf("insufficient funds")
	}
	sender.Money -= amount
	receiver.Money += amount
	return nil
}

func (b *Board) RollDice() int {
	return rand.Intn(12) + 1
}

func (b *Board) HandleAction(player *Player, message Message) (string, string, error) {
	// if not players turn
	switch message.Action {
	case ActionTrade:
		return b.HandleTrade(player, message.Body.(GameTradeBody))
	case ActionAcceptTrade:
		// TODO:
		// Validate and Accept Trade Do Transaction

	case ActionForfeitGame:
		// TODO:
		// remove player from game
		return fmt.Sprintf("%s forfeited the game", player.Name), "", nil
	default:
		fmt.Println("Not Here")
		// if players turn
		if player.Name == b.CurrentPlayer().Name {
			// action is allowed in player's own turn
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
			}
		}
	}
	return "", "", fmt.Errorf("error invalid action")
}

func (b *Board) GetPlayer(id IdType) *Player {
	return b.Players[*id]
}

// Check if trade body is valid
func (b *Board) CheckTradeBody(tradeBody GameTradeBody) error {
	if tradeBody.Id == nil {
		return fmt.Errorf("missing trade id")
	}
	if tradeBody.To == nil || tradeBody.From == nil {
		return fmt.Errorf("missing trade participant")
	}
	return nil
}

// check if player is the owner
func (b *Board) isOwner(player *Player, propertyIds ...IdType) bool {
	for _, i := range propertyIds {
		if b.Slots[*i].Owner != player.Id {
			return false
		}
	}
	return true
}

// Trade Details checker
func (b *Board) CheckTradeDetails(from *Player, tradeBody GameTradeBody) error {
	to := b.GetPlayer(tradeBody.To)
	fmt.Println("name" + to.Name)

	// check if from id is same
	if from.Id != tradeBody.From {
		fmt.Println("from id is not same")
	}

	if tradeBody.Give.Money > from.Money || tradeBody.Take.Money > to.Money {
		return fmt.Errorf("insufficient funds")
	}

	if tradeBody.Give.Money < 0 || tradeBody.Take.Money < 0 {
		return fmt.Errorf("invalid amount")
	}

	if !b.isOwner(from, tradeBody.Give.Property...) {
		return fmt.Errorf("not owner of property")
	}
	if !b.isOwner(to, tradeBody.Take.Property...) {
		return fmt.Errorf("not owner of property")
	}

	return nil
}

// Enlist Trade
func (b *Board) EnlistTrade(tradeBody GameTradeBody) {
	b.Lock()
	b.Trades = append(b.Trades, &tradeBody)
	b.Unlock()
}

func (b *Board) HandleTrade(from *Player, tradeBody GameTradeBody) (string, string, error) {
	err := b.CheckTradeBody(tradeBody)
	if err != nil {
		return "", "", err
	}
	err = b.CheckTradeDetails(from, tradeBody)
	if err != nil {
		return "", "", err
	}

	b.EnlistTrade(tradeBody)
	return fmt.Sprintf("New trade Added, Id: %d", tradeBody.Id), "", nil
}

func (b *Board) BuyProperty(player *Player) (string, string, error) {
	slot := b.Slots[player.Position]
	if slot.Owner != nil {
		return "", "", fmt.Errorf("slot already owned")
	}
	if player.Money < slot.Price {
		return "", "", fmt.Errorf("insufficient funds")
	}
	// TODO:
	// Create a transaction function
	b.TransferPlayerToBank(player, slot.Price)
	slot.Owner = player.Id
	return fmt.Sprintf("%s bought %s for %d", player.Name, slot.Name, slot.Price), "", nil
}

func (b *Board) MovePlayer(player *Player, steps int) (string, string, error) {
	player.Position = (player.Position + steps) % len(b.Slots)
	currentSlot := b.Slots[player.Position]

	switch currentSlot.Type {
	case SlotTypeProperty:
		if currentSlot.Owner == nil && currentSlot.Price > 0 {
			// prompt user
			return "", fmt.Sprintf("Want to buy %s for %d?", currentSlot.Name, currentSlot.Price), nil
		} else if currentSlot.Owner != nil && currentSlot.Owner != player.Id {
			// Pay rent (simple calculation)
			// TODO:
			// proper rent calculation
			rent := currentSlot.Price / 10

			for _, p := range b.Players {
				if p.Id == currentSlot.Owner {
					err := b.TransferPlayerToPlayer(player, p, rent)
					if err != nil {
						return "", "", err
					}
					return fmt.Sprintf("%s paid %d rent to %s", player.Name, rent, currentSlot.Owner), "", nil
				}
			}
		}
	case SlotTypeCard:
		// Handle card slot logic here
		// return "", "", fmt.Errorf("%s", currentSlot.Type)
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

func (b *Board) EndTurn(player *Player) (string, string, error) {
	b.NextTurn()
	return fmt.Sprintf("Waiting for %s to play", b.CurrentPlayer().Name), "", nil
}
