// Package game contains the core game logic and data structures for the Monopoly-like game.
package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TradeHistoryEntry records the details of a completed trade between players.
type TradeHistoryEntry struct {
	RequesterName string
	ResponderName string
	Give          TradeDetails
	Take          TradeDetails
	Timestamp     string
}

// TransferCards transfers cards from one player to another.
func (b *Board) TransferCards(sender *Player, receiver *Player, cards ...IdType) error {
	b.Lock()
	defer b.Unlock()

	for _, card := range cards {
		// Check if the sender has the card
		cardIndex := -1
		for i, c := range sender.Inventory {
			if c == card {
				cardIndex = i
				break
			}
		}

		if cardIndex == -1 {
			return fmt.Errorf("card not found in sender's inventory")
		}

		// Remove the card from the sender's inventory
		sender.Inventory = append(sender.Inventory[:cardIndex], sender.Inventory[cardIndex+1:]...)

		// Add the card to the receiver's inventory
		receiver.Inventory = append(receiver.Inventory, card)
	}

	return nil
}

// Slottype represents the type of a board slot (property, card, jail, etc.).
type Slottype string

const (
	SlotTypeProperty Slottype = "property"
	SlotTypeCard     Slottype = "card"
	SlotTypeJail     Slottype = "jail"
	SlotTypeTax      Slottype = "tax"
	SlotTypeNeutral  Slottype = "neutral"
)

// Slot represents a space on the board (property, card, jail, etc.).
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

// RemovePlayer removes a player from the game and resets their properties.
func (b *Board) RemovePlayer(player *Player) (string, error) {
	b.Lock()
	defer b.Unlock()

	// Find the index of the player to remove
	var index int
	var found bool
	for i, p := range b.Players {
		if p.Id == player.Id {
			index = i
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("player not found")
	}

	// Remove the player from the Players slice
	b.Players = append(b.Players[:index], b.Players[index+1:]...)

	// Handle properties owned by the player
	for i := range b.Slots {
		if b.Slots[i].Owner == player.Id {
			b.Slots[i].Owner = nil
		}
	}

	// Adjust the turn if necessary
	if b.Turn >= len(b.Players) {
		b.Turn = 0
	}

	return fmt.Sprintf("%s has been removed from the game", player.Name), nil
}

// Player represents a player in the game.
type Player struct {
	Id        IdType
	Name      string
	Money     int
	Position  int
	InJail    bool
	JailTurns int
	Inventory []IdType
}

// Board holds the state of the game, including players, slots, trades, and turn management.
type Board struct {
	Slots        []Slot
	Cards        []Card
	Players      []*Player
	Trades       []*GameTradeBody
	TradeHistory []TradeHistoryEntry
	Turn         int
	sync.Mutex
	// Turn Done to prevent player from ending turn without completing Turn Duties
	// After completing a set of actions it unlocks and current players turn can end
	// But still Go/ Move is locekd
	TurnDone bool
	// Move lock to prevent current player to Go multiple times
	MoveLock bool
}

type IdType *int

type (
	// GameTradeBody represents a trade proposal between two players.
	GameTradeBody struct {
		Requester IdType       `json:"requster" validate:"required"`
		Id        IdType       `json:"responder" validate:"required"` // `required` ensures this field must be present
		Responder IdType       `json:"from" validate:"required"`
		Give      TradeDetails `json:"give,omitempty"`
		Take      TradeDetails `json:"take,omitempty"`
		Accept    bool         `json:"accept,omitempty"`
		Active    bool         `json:"active,omitempty"`
	}
	// GameTradeAcceptBody represents a request to accept a trade.
	GameTradeAcceptBody struct {
		TradeId IdType `json:"tradeId" validate:"required"`
	}
)

// TradeDetails describes the assets involved in a trade (properties, money, cards).
type TradeDetails struct {
	Property []IdType `json:"property,omitempty"`
	Money    int      `json:"money,omitempty"`
	Cards    []IdType `json:"cards,omitempty"`
}

// Card represents a special card with an effect in the game.
type Card struct {
	Name        string
	Description string
	Effect      func(*Player, *Board) (string, error)
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
	}
	cards := []Card{
		{Name: "Jail Free Card", Description: "Get out of jail free card", Effect: func(p *Player, b *Board) (string, error) {
			p.InJail = false
			return fmt.Sprintf("%s used a Jail Free Card", p.Name), nil
		}},
		{Name: "Advance to Go", Description: "Advance to Go and collect $200", Effect: func(p *Player, b *Board) (string, error) {
			p.Position = 0
			b.TransferBankToPlayer(p, 200)
			return fmt.Sprintf("%s advanced to Go and collected $200", p.Name), nil
		}},
	}
	return &Board{
		Slots:   slots,
		Cards:   cards,
		Players: []*Player{},
		Turn:    0,
	}
}

// AddPlayer adds a new player to the board and returns the player instance.
func (b *Board) AddPlayer(name string) *Player {
	newId := len(b.Players)
	player := &Player{Name: name, Money: 1500, Position: 0, Id: &newId}
	b.Lock()
	b.Players = append(b.Players, player)
	b.Unlock()
	return player
}

// CurrentPlayer returns the player whose turn it is.
func (b *Board) CurrentPlayer() *Player {
	b.Lock()
	defer b.Unlock()
	return b.Players[b.Turn]
}

// NextTurn advances the turn to the next player.
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
	b.MoveLock = false
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

// TransactionType represents the type of transaction (bank to player, player to bank, etc.).
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

func (b *Board) TransferPlayerToPlayer(sender *Player, receiver *Player, amount int) (func(), error) {
	if sender.Money < amount {
		return nil, fmt.Errorf("insufficient funds")
	}
	sender.Money -= amount
	receiver.Money += amount
	return func() { b.TransferPlayerToPlayer(receiver, sender, amount) }, nil
}

func (b *Board) RollDice() int {
	return rand.Intn(12) + 1
}

// Transfer Properties
func (b *Board) TransferProperty(sender *Player, receiver *Player, properties ...IdType) error {
	b.Lock()
	defer b.Unlock()
	for _, property := range properties {
		// PROBLEM: This will create problem as a serch function will be needed find out propertie's board position.
		if b.Slots[*property].Owner != sender.Id {
			return fmt.Errorf("not owner of property")
		}
	}
	for _, property := range properties {
		b.Slots[*property].Owner = receiver.Id
	}
	return nil
}

// HandleTradeAccept processes a trade acceptance between players.
func (b *Board) HandleTradeAccept(player *Player, tradeAcceptBody GameTradeAcceptBody) (string, string, error) {
	trade := b.Trades[*tradeAcceptBody.TradeId]
	requester := b.GetPlayer(trade.Requester)
	responder := b.GetPlayer(trade.Responder)

	if player.Id != responder.Id {
		return "", "", fmt.Errorf("not your trade")
	}
	var reverts []func()
	revert, err := b.TransferPlayerToPlayer(requester, responder, trade.Give.Money)
	if err != nil {
		return "", "", err
	}

	reverts = append(reverts, revert)

	_, err = b.TransferPlayerToPlayer(responder, requester, trade.Take.Money)
	if err != nil {
		for _, r := range reverts {
			r()
		}
		return "", "", err
	}

	// 2. Transfer Property
	if err := b.TransferProperty(requester, responder, trade.Give.Property...); err != nil {
		return "", "", err
	}

	if err := b.TransferProperty(responder, requester, trade.Take.Property...); err != nil {
		return "", "", err
	}

	// Transfer Cards
	if err := b.TransferCards(requester, responder, trade.Give.Cards...); err != nil {
		return "", "", err
	}

	if err := b.TransferCards(responder, requester, trade.Take.Cards...); err != nil {
		return "", "", err
	}

	return "", "", nil
}

// HandleAction processes a game action from a player.
// This version does NOT depend on room.Message or room.Action* constants.
// Instead, it takes a generic action string and a body (payload), and returns messages/errors.
func (b *Board) HandleAction(player *Player, action string, body interface{}) (string, string, error) {
	// The room package is responsible for interpreting the action string and body.
	switch action {
	case "trade":
		tradeBody, ok := body.(GameTradeBody)
		if !ok {
			return "", "", fmt.Errorf("invalid trade body")
		}
		return b.HandleTrade(player, tradeBody)
	case "accept_trade":
		acceptBody, ok := body.(GameTradeAcceptBody)
		if !ok {
			return "", "", fmt.Errorf("invalid accept trade body")
		}
		return b.HandleTradeAccept(player, acceptBody)
	case "forfeit_game":
		msg, err := b.RemovePlayer(player)
		if err != nil {
			return "", "", err
		}
		return msg, "", nil
	default:
		// Only allow certain actions if it's the player's turn
		if player.Name == b.CurrentPlayer().Name {
			b.LockTurnDone()
			switch action {
			case "go":
				return b.HandleGo(player)
			case "buy":
				return b.BuyProperty(player)
			case "end_turn":
				return b.HandleEndTurn(player)
				// Add more actions as needed
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
	defer b.Unlock()

	// Add the trade to the list of trades
	b.Trades = append(b.Trades, &tradeBody)

	// Create a new trade history entry
	tradeHistoryEntry := TradeHistoryEntry{
		RequesterName: b.GetPlayer(tradeBody.Requester).Name,
		ResponderName: b.GetPlayer(tradeBody.Responder).Name,
		Give:          tradeBody.Give,
		Take:          tradeBody.Take,
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	// Add the trade history entry to the trade history
	b.TradeHistory = append(b.TradeHistory, tradeHistoryEntry)
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

// BuyProperty allows a player to purchase the property they are currently on.
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
func (b *Board) calculateRent(currentSlot Slot) (int, error) {
	if currentSlot.Owner == nil {
		return 0, nil
	}
	if b.Players[*currentSlot.Owner].InJail {
		return 0, nil
	}
	switch currentSlot.State {
	case 0:
		return currentSlot.Rent1, nil
	case 1:
		return currentSlot.Rent2, nil
	case 2:
		return currentSlot.Rent3, nil
	case 3:
		return currentSlot.Rent4, nil
	case 4:
		return currentSlot.Rent5, nil
	default:
		return -1, fmt.Errorf("invalid state")
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

			rent, err := b.calculateRent(currentSlot)
			if err != nil {
				return "", "", err
			}

			for _, p := range b.Players {
				if p.Id == currentSlot.Owner {
					_, err := b.TransferPlayerToPlayer(player, p, rent)
					if err != nil {
						return "", "", err
					}
					return fmt.Sprintf("%s paid %d rent to %d", player.Name, rent, currentSlot.Owner), "", nil
				}
			}
		}
	case SlotTypeCard:
		return b.HandleCardSlot(player, currentSlot)
	case SlotTypeJail:
		return b.HandleJailSlot(player, currentSlot)
	case SlotTypeTax:
		return b.HandleTaxSlot(player, currentSlot)
	case SlotTypeNeutral:
		return b.HandleNeutralSlot(player, currentSlot)
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

	return b.MovePlayer(player, steps)
}

func (b *Board) HandleEndTurn(player *Player) (string, string, error) {
	// TODO: player end turn checks
	// Player cant end turn if
	// 1. he hasnt moved
	// 2. he hasnt bought/auctioned property
	// 3. he hasnt paid rent / balance <0
	// So UnlockTurnDone must be called form
	// Either
	// in jail and Complete Jail Prompt
	// OR
	// 1. No double
	// AND
	// 2. Moved to a slot
	// AND
	// 2. Bought Property
	// OR
	// 2. Paid Rent

	if !b.TurnDone {
		return "", "", fmt.Errorf("turn not done")
	}
	b.NextTurn()
	return fmt.Sprintf("Waiting for %s to play", b.CurrentPlayer().Name), "", nil
}

// Placeholder function for handling card slots
func (b *Board) HandleCardSlot(player *Player, slot Slot) (string, string, error) {
	// TODO: Implement card slot logic
	return "", "", nil
}

// Placeholder function for handling jail slots
func (b *Board) HandleJailSlot(player *Player, slot Slot) (string, string, error) {
	if player.InJail {
		player.JailTurns++
		if player.JailTurns >= 3 {
			player.InJail = false
			player.JailTurns = 0
			b.UnlockPlayerMove(player)
			return fmt.Sprintf("%s is released from jail after serving time", player.Name), "", nil
		}
		return fmt.Sprintf("%s is in jail for %d more turns", player.Name, 3-player.JailTurns), "", nil
	}
	player.InJail = true
	player.JailTurns = 0
	player.Position = b.findJailSlotPosition()
	b.LockPlayerMove(player)
	return fmt.Sprintf("%s has been sent to jail", player.Name), "", nil
}

// Helper function to find the position of the jail slot
func (b *Board) findJailSlotPosition() int {
	for i, slot := range b.Slots {
		if slot.Type == SlotTypeJail {
			return i
		}
	}
	return -1 // Return -1 if no jail slot is found
}

// Function for handling tax slots
func (b *Board) HandleTaxSlot(player *Player, slot Slot) (string, string, error) {
	err := b.TransferPlayerToBank(player, slot.Price)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s paid %d in taxes", player.Name, slot.Price), "", nil
}

// Placeholder function for handling neutral slots

func (b *Board) HandleNeutralSlot(player *Player, slot Slot) (string, string, error) {
	// Return a message indicating the player has landed on a neutral slot
	return fmt.Sprintf("%s has landed on a neutral slot: %s", player.Name, slot.Name), "", nil
}
