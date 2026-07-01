package port

import "time"

type TokenIssuer interface {
	NewToken(subject, role string) (string, time.Time, error)
}
