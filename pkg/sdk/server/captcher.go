package server

import (
	"context"
)

type Challenge struct {
	Id   string
	HTML string
}

type Captcher interface {
	NewChallenge(context.Context, int32) (*Challenge, error)
	HandleFrontendEvent(context.Context, Stream, string, []byte) error
	HandleConnectionClosed(context.Context, string, []byte) error
	HandleBalancerEvent(context.Context, Stream, string, []byte) error
	HandleCloseChallenge(context.Context, Stream, string, []byte) error
	OnStreamStarted(context.Context, Stream) error
	OnStreamClosed(context.Context, Stream)
	Stream(context.Context, Stream)
}

var buildCaptcherFunc captcherBuilder

func SetCaptcherBuilder(b captcherBuilder) {
	if b == nil {
		panic("SetCaptcherBuilder: BuildCaptcherBuilder() is nil")
	}

	if buildCaptcherFunc != nil {
		panic("SetCaptcherBuilder: BuildCaptcherBuilder() is already set elsewhere")
	}

	buildCaptcherFunc = b
}

func GetCaptcherBuilder() captcherBuilder {
	if buildCaptcherFunc == nil {
		panic("GetCaptcherBuilder: BuildCaptcherBuilder() is nil")
	}

	return buildCaptcherFunc
}
