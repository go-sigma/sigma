package dao

type config struct {
	AuditUserID int64
}

// Option ...
type Option func(*config)

// WithAuditUser ...
func WithAuditUser(userID int64) Option {
	return func(c *config) {
		c.AuditUserID = userID
	}
}
