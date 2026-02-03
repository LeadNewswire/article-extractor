package extractor

import "errors"

var (
	// ErrNoContent is returned when no content could be extracted.
	ErrNoContent = errors.New("no content could be extracted")

	// ErrInvalidHTML is returned when the HTML is invalid or cannot be parsed.
	ErrInvalidHTML = errors.New("invalid HTML content")

	// ErrContentTooShort is returned when the extracted content is too short.
	ErrContentTooShort = errors.New("extracted content is too short")

	// ErrHTTPRequest is returned when the HTTP request fails.
	ErrHTTPRequest = errors.New("HTTP request failed")

	// ErrInvalidURL is returned when the URL is invalid.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrContentTooLarge is returned when the content exceeds the maximum size.
	ErrContentTooLarge = errors.New("content exceeds maximum size")

	// ErrTimeout is returned when the operation times out.
	ErrTimeout = errors.New("operation timed out")
)

// ExtractionError wraps an error with additional context.
type ExtractionError struct {
	Op  string // Operation that failed
	URL string // URL being processed (if applicable)
	Err error  // Underlying error
}

func (e *ExtractionError) Error() string {
	if e.URL != "" {
		return e.Op + " [" + e.URL + "]: " + e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

func (e *ExtractionError) Unwrap() error {
	return e.Err
}

// NewExtractionError creates a new ExtractionError.
func NewExtractionError(op string, url string, err error) *ExtractionError {
	return &ExtractionError{
		Op:  op,
		URL: url,
		Err: err,
	}
}
