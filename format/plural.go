package format

import (
	"log/slog"
	"sync"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var registerMessagesOnce sync.Once

const (
	Days        = "%d day"
	Movies      = "%d movie"
	Nominations = "%d nomination"
	Users       = "%d user"
	Wins        = "%d win"
)

func initMessages() {
	register(Days, "%d day", "%d days")
	register(Movies, "%d movie", "%d movies")
	register(Nominations, "%d nomination", "%d nominations")
	register(Wins, "%d win", "%d wins")
	register(Users, "%d user", "%d users")
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

// Plural formats a plural message with the given count and key.
// It ensures messages are registered only once.
func Plural(key string, count int) string {
	registerMessagesOnce.Do(initMessages)

	p := message.NewPrinter(language.English)
	return p.Sprintf(key, count)
}
