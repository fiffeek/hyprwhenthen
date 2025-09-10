package hypr

import (
	"strings"

	"github.com/fiffeek/hyprwhenthen/internal/config"
)

type Event struct {
	EventType         string
	EventContext      string
	EventContextBytes []byte
}

func getRegisteredEvent(cfg *config.RawConfig, line string) (bool, *Event) {
	for _, key := range cfg.EventKeys {
		after, done := strings.CutPrefix(line, key+">>")
		if done {
			return true, &Event{
				EventType:         key,
				EventContext:      after,
				EventContextBytes: []byte(after),
			}
		}
	}
	return false, nil
}
