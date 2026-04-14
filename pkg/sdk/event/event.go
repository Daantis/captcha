package event

const (
	ServerEventTypeUnset = iota
	ServerEventTypeChallengeResult
	ServerEventTypeClientData
)

const (
	ClientEventTypeUnset = iota
	ClientEventFrontendEvent
	ClientEventTypeConnectionClosed
	ClientEventTypeBalancerEvent
	ClientEventTypeCancelChallenge
)

type ClientEvent struct {
	EventType   int
	ChallengeId string
	Data        []byte
}

type ServerEvent struct {
	EventType   int
	ChallengeId string
	Event       any
}

type ServerEventChallengeResult struct {
	ConfidencePercent int32
}

type ServerEventClientData struct {
	Data []byte
}
