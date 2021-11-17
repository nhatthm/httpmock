package value

// ErrUnsupportedDataType represents that the data type is not supported.
const ErrUnsupportedDataType err = "unsupported data type"

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}
