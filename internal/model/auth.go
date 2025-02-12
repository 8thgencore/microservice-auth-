package model

// UserCreds type is the structure for user sign in.
type UserCreds struct {
	Username string
	Password string
}

// AuthInfo type is the structure for user authentication data from storage.
type AuthInfo struct {
	ID       string
	Username string
	Password string
	Role     string
	Version  int
}

// TokenPair type is the structure for storing access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
