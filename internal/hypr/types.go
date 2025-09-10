package hypr

import (
	"hyprwhenthen/internal/config"
	"strings"
)

type Event struct {
	EventType    string
	EventContext string
}

func ParseEvent(cfg *config.RawConfig, line string) (bool, *Event) {
	for _, key := range cfg.EventKeys {
		after, done := strings.CutPrefix(line, key+">>")
		if done {
			return true, &Event{
				EventType:    key,
				EventContext: after,
			}
		}
	}
	return false, nil
}
