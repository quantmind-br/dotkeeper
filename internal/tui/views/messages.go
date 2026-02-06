package views

// ErrorMsg represents a generic error message that can be used across views.
// The Source field identifies which view or operation generated the error.
type ErrorMsg struct {
	Source string
	Err    error
}

// SuccessMsg represents a generic success message that can be used across views.
// The Source field identifies which view or operation succeeded.
type SuccessMsg struct {
	Source  string
	Message string
}

// LoadingMsg represents a loading state message that can be used across views.
// The Source field identifies which view or operation is loading.
// The Message field provides context about what is being loaded.
type LoadingMsg struct {
	Source  string
	Message string
}

// RefreshMsg represents a refresh request message that can be used across views.
// The Source field identifies which view should refresh (empty string = all views).
type RefreshMsg struct {
	Source string
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
