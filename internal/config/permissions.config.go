package config

const (
	// PermissionProfileRead grants reading any profile by its ID, and the caller's own.
	PermissionProfileRead = "profile:read"
	// PermissionProfileWrite grants changing the caller's own profile.
	PermissionProfileWrite = "profile:write"
)
