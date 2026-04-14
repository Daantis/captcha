package client

import (
	"context"
	"sdk/event"
)

type Stream interface {
	Send(event *event.ClientEvent) error
	Recv() (*event.ServerEvent, error)
	CloseSend() error
	Context() context.Context
}
