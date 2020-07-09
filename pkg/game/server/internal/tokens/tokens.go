package tokens

const (
	// X token
	X = "X"
	// O token
	O = "O"
)

// FromIndex converts an integer index to a token in a predictable way.
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
