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
	Name  string
	Type  Slottype
	Owner IdType
	Price int
	State int
	Rent1 int
	Rent2 int
	Rent3 int
	Rent4 int
	Rent5 int
	// appends slot here
	// Slot is a slot
	// slot is kind of like parent class
}

type Player struct {
	Id        IdType
	Name      string
	Money     int
	Position  int
	InJail    bool
	Inventory []IdType
}

type Board struct {
	Slots   []Slot
	Cards   []Card
	Players []*Player
	Trades  []*GameTradeBody
	Turn    int
	sync.Mutex
	// Turn Done to prevent player from ending turn without completing Turn Duties
	// After completing a set of actions it unlocks and current players turn can end
	// But still Go/ Move is locekd
	// If player does end turn it unlocks MoveLock
	TurnDone bool
	// Move lock to prevent current player to Go multiple times
	MoveLock bool
}

type IdType *int

type (
	GameTradeBody struct {
		Requester IdType       `json:"requster" validate:"required"`
		Id        IdType       `json:"responder" validate:"required"` // `required` ensures this field must be present
		Responder IdType       `json:"from" validate:"required"`
		Give      TradeDetails `json:"give,omitempty"`
		Take      TradeDetails `json:"take,omitempty"`
		Accept    bool         `json:"accept,omitempty"`
		Active    bool         `json:"active,omitempty"`
	}
	GameTradeAcceptBody struct {
		TradeId IdType `json:"tradeId" validate:"required"`
	}
)

type TradeDetails struct {
	Property []IdType `json:"property,omitempty"`
	Money    int      `json:"money,omitempty"`
	Cards    []IdType `json:"cards,omitempty"`
}

type Card struct {
	Name        string
	Description string
}

// LOOC YREV
// ----------
func NewBoard() *Board {
	// Create a simple board with properties

	slots := []Slot{
		{Name: "Mediterranean Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		{Name: "Arkochan Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		{Name: "Chittagong", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		// {Name: "Community Chest", Type: SlotTypeCard, Owner: nil, Price: 0, Houses: 0},
		{Name: "Hell Yeah Avenue", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		{Name: "Nicsu York", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		{Name: "MiniSoda", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		{Name: "Ohio", Type: SlotTypeProperty, Owner: nil, Price: 60, State: 0},
		// {Name: "Income Tax", Type: SlotTypeTax, Owner: nil, Price: 0, Houses: 0},
		// {Name: "Go to Jail", Type: SlotTypeJail, Owner: nil, Price: 0, Houses: 0},
		// {Name: "Free Parking", Type: SlotTypeNeutral, Owner: nil, Price: 0, Houses: 0},
	}
	cards := []Card{
		{Name: "Jail Free Card", Description: "Get out of jail free card"},
	}
	return &Board{
		Slots:   slots,
		Cards:   cards,
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

// lock and unlock player movelock property
func (b *Board) LockPlayerMove(player *Player) {
	b.Lock()
	b.MoveLock = true
	b.Unlock()
}

func (b *Board) UnlockPlayerMove(player *Player) {
	b.Lock()
	b.MoveLock = true
	b.Unlock()
}

// lock and unlock TurnDone property
func (b *Board) LockTurnDone() {
	b.Lock()
	b.TurnDone = true
	b.Unlock()
}

func (b *Board) UnlockTurnDone() {
	b.Lock()
	b.TurnDone = false
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

// Transfer Properties
func (b *Board) TransferProperty(sender *Player, receiver *Player, properties ...IdType) error {
	b.Lock()
	defer b.Unlock()
	for _, property := range properties {
		if b.Slots[*property].Owner != sender.Id {
			return fmt.Errorf("not owner of property")
		}
	}
	for _, property := range properties {
		b.Slots[*property].Owner = receiver.Id
	}
	return nil
}

// TradeAcceptHandler
func (b *Board) HandleTradeAccept(player *Player, tradeAcceptBody GameTradeAcceptBody) (string, string, error) {
	trade := b.Trades[*tradeAcceptBody.TradeId]
	requester := b.GetPlayer(trade.Requester)
	responder := b.GetPlayer(trade.Responder)

	if player.Id != responder.Id {
		return "", "", fmt.Errorf("not your trade")
	}

	if err := b.TransferPlayerToPlayer(requester, responder, trade.Give.Money); err != nil {
		return "", "", err
	}
	if err := b.TransferPlayerToPlayer(responder, requester, trade.Take.Money); err != nil {
		return "", "", err
	}

	// 2. Transfer Property
	if err := b.TransferProperty(requester, responder, trade.Give.Property...); err != nil {
		return "", "", err
	}
	if err := b.TransferProperty(responder, requester, trade.Take.Property...); err != nil {
		return "", "", err
	}

	// TODO: 3. Transfer Cards
	return "", "", nil
}

func (b *Board) HandleAction(player *Player, message Message) (string, string, error) {
	// if not players turn
	switch message.Action {
	case ActionTrade:
		return b.HandleTrade(player, message.Body.(GameTradeBody))
	case ActionAcceptTrade:
		return b.HandleTradeAccept(player, message.Body.(GameTradeAcceptBody))
	case ActionForfeitGame:
		// TODO:
		// remove player from game
		return fmt.Sprintf("%s forfeited the game", player.Name), "", nil
	default:
		fmt.Println("Not Here")
		// if players turn
		if player.Name == b.CurrentPlayer().Name {
			// action is allowed in player's own turn
			b.LockTurnDone()
			switch message.Action {
			case ActionGo:
				return b.HandleGo(player)
			case ActionBuy:
				return b.BuyProperty(player)
			case ActionEndTurn:
				return b.HandleEndTurn(player)

				// TODO: Actions
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
	if tradeBody.Requester == nil || tradeBody.Responder == nil {
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
	to := b.GetPlayer(tradeBody.Requester)
	fmt.Println("name" + to.Name)

	// check if from id is same
	if from.Id != tradeBody.Responder {
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
	b.TransferPlayerToBank(player, slot.Price)
	slot.Owner = player.Id
	return fmt.Sprintf("%s bought %s for %d", player.Name, slot.Name, slot.Price), "", nil
}

// function to calculate rent
func (b *Board) calculateRent(currentSlot Slot) int {
	if currentSlot.Owner == nil {
		return 0
	}
	if b.Players[*currentSlot.Owner].InJail {
		return 0
	}
	switch currentSlot.State {
	case 0:
		return currentSlot.Rent1
	case 1:
		return currentSlot.Rent2
	case 2:
		return currentSlot.Rent3
	case 3:
		return currentSlot.Rent4
	case 4:
		return currentSlot.Rent5
	default:
		return -1
	}
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

			rent := b.calculateRent(currentSlot)
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

func (b *Board) HandleGo(player *Player) (string, string, error) {
	if b.MoveLock {
		return "", "", fmt.Errorf("cant go now")
	}
	b.LockPlayerMove(player)

	// Player moving shouldn't be unlocked after move
	// Only after the player has finished his turn
	// So End turn Should have checks
	// defer b.UnlockPlayerMove(player)

	steps := b.RollDice()
	// TODO:
	// Handle Double

	if player.InJail {
		// TODO: Jail double roll
		// If in jail go should lead here
		// Roll dice
		// if double
		// Out
		return "", "", nil
	}
	return b.MovePlayer(player, steps)
}

func (b *Board) HandleEndTurn(player *Player) (string, string, error) {
	// TODO: player end turn checks
	// Player cant end turn if
	// 1. he hasnt moved
	// 2. he hasnt bought/auctioned property
	// 3. he hasnt paid rent / balance <0

	if !b.TurnDone {
		return "", "", fmt.Errorf("turn not done")
	}
	b.NextTurn()
	return fmt.Sprintf("Waiting for %s to play", b.CurrentPlayer().Name), "", nil
}
