package types

type userContextKey string

// UserInfoKey is the key used to store the user information in the request context.
const UserInfoKey userContextKey = "user-info"

// UserInfo represents the information about a user.
type UserInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

// APIRoot represents environment data for the Cluster Manager.
type APIRoot struct {
	UserInfo
	Trusted bool `json:"trusted"`
}
