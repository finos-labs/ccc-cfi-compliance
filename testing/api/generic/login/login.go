package login

// Login refreshes cloud CLI / default credentials when the current token is near expiry
// (JWT exp, Azure expires_on, AWS_CREDENTIAL_EXPIRATION), not a fixed wall-clock interval.
type Login interface {
	EnsureLoginToken(cloud string) error
}

// Default is the shared Login used by the factory and policy checker.
var Default Login = NewManager()
