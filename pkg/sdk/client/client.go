package client

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	v1 "sdk/pkg/pb/v1"
	"time"
)

type Client struct {
	api   v1.CaptchaServiceClient
	close func() error
}

func New(_ context.Context, target string) (*Client, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                2 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: false,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create grpc client")
	}

	return &Client{
		api: v1.NewCaptchaServiceClient(conn),
		close: func() error {
			return conn.Close()
		},
	}, nil
}

type NewChallengeRequest struct {
	Complexity int32 `json:"complexity"`
}

type NewChallengeResponse struct {
	ChallengeId string `json:"challenge_id"`
	HTML        string `json:"html"`
}

func (c *Client) NewChallenge(ctx context.Context, request *NewChallengeRequest) (*NewChallengeResponse, error) {
	if c.api == nil {
		return nil, errors.New("api is nil")
	}

	resp, err := c.api.NewChallenge(ctx, &v1.ChallengeRequest{
		Complexity: request.Complexity,
	})
	if err != nil {
		return nil, errors.Wrap(err, "new challenge failed")
	}

	// todo(nth): validate

	return &NewChallengeResponse{
		ChallengeId: resp.ChallengeId,
		HTML:        resp.Html,
	}, nil
}

func (c *Client) MakeEventStream(ctx context.Context) (Stream, error) {
	grpcStream, err := c.api.MakeEventStream(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "make event stream failed")
	}

	return &stream{grpcStream: grpcStream}, nil
}

func (c *Client) Close() error {
	return c.close()
}
