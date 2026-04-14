package client

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"sdk/event"
	v1 "sdk/pkg/pb/v1"
)

type stream struct {
	grpcStream v1.CaptchaService_MakeEventStreamClient
}

func (s *stream) Send(event *event.ClientEvent) error {
	if _, ok := clientEventTypes[event.EventType]; !ok {
		return errors.New("unknown client event type")
	}

	if len(event.ChallengeId) == 0 {
		return errors.New("challenge id is empty")
	}

	err := s.grpcStream.Send(&v1.ClientEvent{
		EventType:   clientEventTypes[event.EventType],
		ChallengeId: event.ChallengeId,
		Data:        event.Data,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send clientEvent")
	}

	return nil
}

func (s *stream) Recv() (*event.ServerEvent, error) {
	res, err := s.grpcStream.Recv()
	if err == io.EOF {
		return nil, io.EOF
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to receive server event")
	}

	var eventData any
	var eventType int

	switch res.Event.(type) {
	case *v1.ServerEvent_Result:
		resEvent := res.Event.(*v1.ServerEvent_Result)
		eventType = event.ServerEventTypeChallengeResult
		eventData = &event.ServerEventChallengeResult{
			ConfidencePercent: resEvent.Result.ConfidencePercent,
		}

	case *v1.ServerEvent_ClientData:
		resEvent := res.Event.(*v1.ServerEvent_ClientData)
		eventType = event.ServerEventTypeClientData
		eventData = &event.ServerEventClientData{
			Data: resEvent.ClientData.Data,
		}

	default:
		return nil, errors.New("invalid server event")
	}

	return &event.ServerEvent{
		EventType:   eventType,
		ChallengeId: res.ChallengeId,
		Event:       eventData,
	}, nil
}

func (s *stream) CloseSend() error {
	err := s.grpcStream.CloseSend()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}

	return nil
}

func (s *stream) Context() context.Context {
	return s.grpcStream.Context()
}

var clientEventTypes = map[int]v1.ClientEvent_EventType{
	event.ClientEventFrontendEvent:        v1.ClientEvent_FRONTEND_EVENT,
	event.ClientEventTypeConnectionClosed: v1.ClientEvent_CONNECTION_CLOSED,
	event.ClientEventTypeBalancerEvent:    v1.ClientEvent_BALANCER_EVENT,
	event.ClientEventTypeCancelChallenge:  v1.ClientEvent_ClOSE_CHALLENGE_EVENT,
}
