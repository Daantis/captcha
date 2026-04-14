package captcha

import "testing"

func TestClientFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := ClientFrame{
		Opcode:  ClientOpDragDrop,
		Seq:     42,
		Phase:   3,
		Subject: 11,
		Target:  17,
		Value:   -1,
		Extra:   88,
	}

	decoded, err := DecodeClientEvent(EncodeClientFrame(frame))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded != frame {
		t.Fatalf("frame mismatch: got %#v want %#v", decoded, frame)
	}
}

func TestServerFrameRoundTripWithPrefix(t *testing.T) {
	t.Parallel()

	frame := ServerFrame{
		Opcode:   ServerOpPatch,
		Seq:      7,
		Phase:    2,
		EntityID: 19,
		Progress: 55,
		Flags:    1,
		Payload:  MarshalPayload(ServerPayload{Message: "ok", View: ViewModel{Mode: "odd-grid", Theme: "odd-grid", Title: "title", Instruction: "inst", ProgressText: "1/1", Layout: LayoutModel{Type: "grid"}}}),
	}

	data, err := EncodeServerFrame(frame)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	prefixed := append([]byte{0x80}, data...)
	decoded, err := DecodeServerFrame(prefixed)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Opcode != frame.Opcode || decoded.Seq != frame.Seq || decoded.Phase != frame.Phase {
		t.Fatalf("header mismatch: got %#v want %#v", decoded, frame)
	}
}
