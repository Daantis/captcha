package captcha

import (
	"context"
	"testing"

	"sdk/event"
	"sdk/server"
)

type mockStream struct {
	ctx         context.Context
	clientData  [][]byte
	confidence  []int32
}

func (m *mockStream) SendChallengeResult(_ string, result *event.ServerEventChallengeResult) error {
	m.confidence = append(m.confidence, result.ConfidencePercent)
	return nil
}

func (m *mockStream) SendClientData(_ string, payload *event.ServerEventClientData) error {
	m.clientData = append(m.clientData, append([]byte(nil), payload.Data...))
	return nil
}

func (m *mockStream) Recv() (*event.ClientEvent, error) {
	return nil, context.Canceled
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func TestCaptchaResendsPromptOnStaleSequence(t *testing.T) {
	t.Parallel()

	c, err := NewCaptcha(server.ChallengeId("odd-grid"))
	if err != nil {
		t.Fatalf("new captcha: %v", err)
	}

	challenge, err := c.NewChallenge(context.Background(), 20)
	if err != nil {
		t.Fatalf("new challenge: %v", err)
	}

	stream := &mockStream{ctx: context.Background()}
	if err := c.HandleFrontendEvent(context.Background(), stream, challenge.Id, EncodeClientFrame(ClientFrame{Opcode: ClientOpReady, Seq: 1})); err != nil {
		t.Fatalf("ready: %v", err)
	}
	if len(stream.clientData) == 0 {
		t.Fatal("expected initial frame")
	}
	frame, err := DecodeServerFrame(stream.clientData[0])
	if err != nil {
		t.Fatalf("decode initial frame: %v", err)
	}
	if frame.EntityID == 0 {
		t.Fatal("expected non-zero server entity id")
	}

	session, ok := c.store.Get(challenge.Id)
	if !ok {
		t.Fatal("session not found")
	}

	session.mu.Lock()
	phase := session.Phase
	staleKey := session.ModeState.(*oddGridState).Rounds[0].Correct
	subject := lookupTokenID(session, staleKey)
	session.mu.Unlock()

	if err := c.HandleFrontendEvent(context.Background(), stream, challenge.Id, EncodeClientFrame(ClientFrame{
		Opcode:  ClientOpTap,
		Seq:     1,
		Phase:   phase,
		Subject: subject,
	})); err != nil {
		t.Fatalf("stale event: %v", err)
	}

	if len(stream.clientData) < 2 {
		t.Fatal("expected prompt frame after stale event")
	}
}
