package model

type WSEvent string

const (
	WSEventLOGI WSEvent = "LOGI"
	WSEventMESG WSEvent = "MESG"
	WSEventJOIN WSEvent = "JOIN"
	WSEventPING WSEvent = "PING"
	WSEventPONG WSEvent = "PONG"
	WSEventEROR WSEvent = "EROR"
)

var validWSEvents = map[WSEvent]struct{}{
	WSEventLOGI: {},
	WSEventMESG: {},
	WSEventJOIN: {},
	WSEventPING: {},
	WSEventPONG: {},
	WSEventEROR: {},
}

func IsValidWSEvent(event string) bool {
	if len(event) != 4 {
		return false
	}
	_, ok := validWSEvents[WSEvent(event)]
	return ok
}
