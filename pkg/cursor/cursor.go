package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Cursor represents the decoded pagination cursor
type Cursor struct {
	Timestamp int64     `json:"t"`  // Unix Microseconds
	ID        uuid.UUID `json:"id"` // Unique ID for tie-breaking
}

// Encode creates an opaque cursor string from a time and ID
func Encode(t time.Time, id uuid.UUID) string {
	c := Cursor{
		Timestamp: t.UnixMicro(),
		ID:        id,
	}
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses an opaque cursor string back into time and ID
func Decode(token string) (time.Time, uuid.UUID, error) {
	if token == "" {
		return time.Time{}, uuid.Nil, errors.New("empty cursor token")
	}

	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return time.Time{}, uuid.Nil, err
	}

	return time.UnixMicro(c.Timestamp), c.ID, nil
}
