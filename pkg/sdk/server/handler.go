package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sdk/event"
	v1 "sdk/pkg/pb/v1"
)

func mustBuildHandler(c Captcher, isc incomingStreamCloser) *handler {
	if c == nil {
		panic(errors.New("captcher is nil"))
	}

	if isc == nil {
		panic(errors.New("closer is nil"))
	}

	return &handler{
		captcher:             c,
		incomingStreamCloser: isc,
	}
}

type handler struct {
	v1.UnimplementedCaptchaServiceServer

	captcher             Captcher
	incomingStreamCloser incomingStreamCloser
}

func (s *handler) NewChallenge(ctx context.Context, req *v1.ChallengeRequest) (*v1.ChallengeResponse, error) {
	if s.captcher == nil {
		return nil, errors.New("captcher not set")
	}

	resp, err := s.captcher.NewChallenge(ctx, req.Complexity)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("response is nil")
	}

	return &v1.ChallengeResponse{
		ChallengeId: resp.Id,
		Html:        resp.HTML,
	}, nil
}
func (s *handler) MakeEventStream(grpcStream v1.CaptchaService_MakeEventStreamServer) error {
	eventStream := newStream(grpcStream)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer s.captcher.OnStreamClosed(ctx, eventStream)

	err := s.captcher.OnStreamStarted(ctx, eventStream)
	if err != nil {
		return err
	}

	go s.captcher.Stream(ctx, eventStream)

	for {
		req, err := eventStream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Debug("stream closed eof")
			break
		}

		if err != nil {
			slog.Error(err.Error())
			break
		}

		switch req.EventType {
		case event.ClientEventFrontendEvent:
			err = s.captcher.HandleFrontendEvent(ctx, eventStream, req.ChallengeId, req.Data)
			if err != nil {
				slog.Error(err.Error())
			}

		case event.ClientEventTypeConnectionClosed:
			err = s.incomingStreamCloser.closeIncomingStream()
			if err != nil {
				slog.Error(err.Error())
			}

			err = s.captcher.HandleConnectionClosed(ctx, req.ChallengeId, req.Data)
			if err != nil {
				slog.Error(err.Error())
			}
			break

		case event.ClientEventTypeBalancerEvent:
			err = s.captcher.HandleBalancerEvent(ctx, eventStream, req.ChallengeId, req.Data)
			if err != nil {
				slog.Error(err.Error())
			}

		case event.ClientEventTypeCancelChallenge:
			err = s.captcher.HandleCloseChallenge(ctx, eventStream, req.ChallengeId, req.Data)
			if err != nil {
				slog.Error(err.Error())
			}

		default:
			slog.Error("unknown client event type")
		}
	}

	return nil
}
