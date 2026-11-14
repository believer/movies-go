package auth

import "context"

// Authenticated reports whether the request's user is logged in, as set by the
// app's middleware. A missing value counts as logged out, so components render
// in tests without it.
func Authenticated(ctx context.Context) bool {
	authenticated, _ := ctx.Value("IsAuthenticated").(bool)

	return authenticated
}
