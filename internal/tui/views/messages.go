package views

// ErrorMsg represents a generic error message that can be used across views.
// The Source field identifies which view or operation generated the error.
type ErrorMsg struct {
	Source string
	Err    error
}

// Error returns the error string, implementing the error interface.
func (e ErrorMsg) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying error for errors.Is/errors.As compatibility.
func (e ErrorMsg) Unwrap() error {
	return e.Err
}
