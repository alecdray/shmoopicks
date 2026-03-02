package sqlx

import (
	"database/sql"
	"fmt"
	"time"
)

func NewNullTime(t *time.Time) sql.NullTime {
	if t == nil || t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func DurationToSQLiteDatetime(d time.Duration) string {
	return fmt.Sprintf("-%d seconds", int(d.Seconds()))
}
