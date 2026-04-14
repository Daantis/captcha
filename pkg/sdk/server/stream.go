package server

import (
	"context"
	"github.com/pkg/errors"
	"sdk/event"
	v1 "sdk/pkg/pb/v1"
)

type Stream interface {
	SendChallengeResult(challengeId string, event *event.ServerEventChallengeResult) error
	SendClientData(challengeId string, event *event.ServerEventClientData) error
	Recv() (*event.ClientEvent, error)
	Context() context.Context
}

func newStream(s v1.CaptchaService_MakeEventStreamServer) *stream {
	return &stream{
		grpcStream: s,
	}
}

type stream struct {
	grpcStream v1.CaptchaService_MakeEventStreamServer
}

func (s *stream) SendChallengeResult(challengeId string, event *event.ServerEventChallengeResult) error {
	// todo(nth): validate

	err := s.grpcStream.Send(&v1.ServerEvent{
		Event: &v1.ServerEvent_Result{
			Result: &v1.ServerEvent_ChallengeResult{
				ConfidencePercent: event.ConfidencePercent,
			},
		},
		ChallengeId: challengeId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send clientEvent")
	}

	return nil
}

func (s *stream) SendClientData(challengeId string, event *event.ServerEventClientData) error {
	// todo(nth): validate

	err := s.grpcStream.Send(&v1.ServerEvent{
		Event: &v1.ServerEvent_ClientData{
			ClientData: &v1.ServerEvent_SendClientData{
				Data: event.Data,
			},
		},
		ChallengeId: challengeId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send clientEvent")
	}

	return nil
}

func (s *stream) Recv() (*event.ClientEvent, error) {
	res, err := s.grpcStream.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "failed to receive client event")
	}

	return &event.ClientEvent{
		EventType:   parseClientEventType(res.EventType),
		ChallengeId: res.ChallengeId,
		Data:        res.Data,
	}, nil

}

func (s *stream) Context() context.Context {
	return s.grpcStream.Context()
}

func parseClientEventType(eventType v1.ClientEvent_EventType) int {
	et, ok := clientEventTypes[eventType]
	if !ok {
		return event.ClientEventTypeUnset
	}

	return et
}

var clientEventTypes = map[v1.ClientEvent_EventType]int{
	v1.ClientEvent_FRONTEND_EVENT:        event.ClientEventFrontendEvent,
	v1.ClientEvent_CONNECTION_CLOSED:     event.ClientEventTypeConnectionClosed,
	v1.ClientEvent_BALANCER_EVENT:        event.ClientEventTypeBalancerEvent,
	v1.ClientEvent_ClOSE_CHALLENGE_EVENT: event.ClientEventTypeCancelChallenge,
}
