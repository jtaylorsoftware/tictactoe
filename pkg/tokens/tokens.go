package tokens

const (
	// X token
	X = "X"
	// O token
	O = "O"
	// Empty represents an empty space
	Empty = "_"
)

// FromIndex assigns a token value ("X" or "O") to an index, such that a player
// in the lobby receives a token based on their index in the lobby, which remains in use
// until they disconnect.
func FromIndex(index int) string {
	switch index {
	case 0:
		return X
	case 1:
		return O
	default:
		return ""
	}
}
