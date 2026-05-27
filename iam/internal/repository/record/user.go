package record

import "time"

type User struct {
	UUID         string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
