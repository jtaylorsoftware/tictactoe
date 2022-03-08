package player

import (
	"errors"
	"github.com/jeremyt135/tictactoe/pkg/server/internal/config"
)

// Array provides methods for manipulating an array of Players.
type Array interface {
	At(i int) *Player
	Add(p *Player) (int, error)
	Remove(i int)
	Size() int
	IsFull() bool
}

// FixedArray implements the PlayerArray. It is fixed size but keeps track of how many
// players it actually holds.
type FixedArray struct {
	players [config.MaxPlayers]*Player
	count   int
}

// NewFixedArray returns a FixedArray struct implementing the Array interface.
func NewFixedArray() Array {
	return &FixedArray{}
}

func (arr *FixedArray) nextIndex() (ind int) {
	ind = -1
	for i, p := range arr.players {
		if p == nil {
			ind = i
			return
		}
	}
	return
}

// At returns the Player data for the array cell at index i.
func (arr *FixedArray) At(i int) *Player {
	return arr.players[i]
}

// Add stores Player data at the next available slot, returning
// the index the player was added at.
//
// If the array is full, Add returns an error.
func (arr *FixedArray) Add(p *Player) (int, error) {
	if arr.isFullInternal() {
		return -1, errors.New("playerArr is full")
	}
	ind := arr.nextIndex()
	arr.players[ind] = p
	arr.count++
	return ind, nil
}

// Remove discards the stored Player data at index i.
func (arr *FixedArray) Remove(i int) {
	arr.players[i] = nil
	arr.count--
}

func (arr *FixedArray) isFullInternal() bool {
	return arr.count == config.MaxPlayers
}

// IsFull returns true if the FixedArray contains the max allowed players.
func (arr *FixedArray) IsFull() bool {
	return arr.isFullInternal()
}

// Size returns the number of players actually stored in the FixedArray.
func (arr *FixedArray) Size() int {
	return arr.count
}
