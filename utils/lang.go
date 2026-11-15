package utils

import (
	"log/slog"
	"sync"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var registerMessagesOnce sync.Once

const (
	NominationKey = "%d nomination"
	UsersKey      = "%d user"
	WinKey        = "%d win"
)

func initMessages() {
	register(NominationKey, "%d nomination", "%d nominations")
	register(WinKey, "%d win", "%d wins")
	register(UsersKey, "%d user", "%d users")
}

func register(key, one, other string) {
	msg := plural.Selectf(1, "%d",
		plural.One, one,
		plural.Other, other,
	)

	err := message.Set(language.English, key, msg)

	if err != nil {
		slog.Error("Could not register plural message", "key", key, "error", err)
	}
}

// PluralMessage formats a plural message with the given count and key.
// It ensures messages are registered only once.
func PluralMessage(key string, count int) string {
	registerMessagesOnce.Do(initMessages)

	p := message.NewPrinter(language.English)
	return p.Sprintf(key, count)
}
