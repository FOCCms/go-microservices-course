package redis_view

type Session struct {
	UUID        string `redis:"uuid"`
	UserUUID    string `redis:"user_uuid"`
	Login       string `redis:"login"`
	CreatedAtNs int64  `redis:"created_at"`
	UpdatedAtNs *int64 `redis:"updated_at,omitempty"`
	ExpiresAtNs int64  `redis:"expires_at"`
}
