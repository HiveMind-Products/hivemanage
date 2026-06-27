package auth

import "errors"

// ErrDiscordAlreadyLinked is returned when the Discord account a user is trying
// to link is already attached to a different local account.
var ErrDiscordAlreadyLinked = errors.New("this Discord account is already linked to another user")

// ErrUnlinkWouldLockOut is returned when unlinking Discord would leave an account
// with no way to log in (no password set).
var ErrUnlinkWouldLockOut = errors.New("set a password before unlinking Discord")

type ErrSessionExpired struct{}

func (ErrSessionExpired) Error() string {
	return "auth session expired"
}

func (ErrSessionExpired) Is(target error) bool {
	_, ok := target.(ErrSessionExpired)
	return ok
}

type ErrUserCredentials struct{}

func (ErrUserCredentials) Error() string {
	return "username or password is wrong"
}

func (ErrUserCredentials) Is(target error) bool {
	_, ok := target.(ErrUserCredentials)
	return ok
}
