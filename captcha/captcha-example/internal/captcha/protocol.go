package captcha

import (
	"encoding/binary"
	"encoding/json"
	"errors"
)

const (
	ClientOpReady    byte = 1
	ClientOpTap      byte = 2
	ClientOpSwipe    byte = 3
	ClientOpDragDrop byte = 4
	ClientOpAckFrame byte = 5
)

const (
	ServerOpInit     byte = 101
	ServerOpPatch    byte = 102
	ServerOpPrompt   byte = 103
	ServerOpProgress byte = 104
	ServerOpResult   byte = 105
)

const (
	clientFrameSize = 12
	serverHeaderLen = 10
)

type ClientFrame struct {
	Opcode  byte
	Seq     uint16
	Phase   uint8
	Subject uint16
	Target  uint16
	Value   int16
	Extra   int16
}

func (f ClientFrame) ValueU16() uint16 {
	return uint16(f.Value)
}

func EncodeClientFrame(frame ClientFrame) []byte {
	buf := make([]byte, clientFrameSize)
	buf[0] = frame.Opcode
	binary.LittleEndian.PutUint16(buf[1:3], frame.Seq)
	buf[3] = frame.Phase
	binary.LittleEndian.PutUint16(buf[4:6], frame.Subject)
	binary.LittleEndian.PutUint16(buf[6:8], frame.Target)
	binary.LittleEndian.PutUint16(buf[8:10], uint16(frame.Value))
	binary.LittleEndian.PutUint16(buf[10:12], uint16(frame.Extra))
	return buf
}

func DecodeClientEvent(data []byte) (ClientFrame, error) {
	payload := stripTransportPrefix(data)
	if len(payload) < clientFrameSize {
		return ClientFrame{}, errors.New("client frame is too short")
	}

	return ClientFrame{
		Opcode:  payload[0],
		Seq:     binary.LittleEndian.Uint16(payload[1:3]),
		Phase:   payload[3],
		Subject: binary.LittleEndian.Uint16(payload[4:6]),
		Target:  binary.LittleEndian.Uint16(payload[6:8]),
		Value:   int16(binary.LittleEndian.Uint16(payload[8:10])),
		Extra:   int16(binary.LittleEndian.Uint16(payload[10:12])),
	}, nil
}

type ServerFrame struct {
	Opcode   byte            `json:"opcode"`
	Seq      uint16          `json:"seq"`
	Phase    uint8           `json:"phase"`
	EntityID uint16          `json:"entityId"`
	Progress uint8           `json:"progress"`
	Flags    uint8           `json:"flags"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}

type ServerPayload struct {
	Message string    `json:"message,omitempty"`
	View    ViewModel `json:"view"`
}

type ServerAction struct {
	Frame ServerFrame
}

func EncodeServerFrame(frame ServerFrame) ([]byte, error) {
	payload := frame.Payload
	if payload == nil {
		payload = []byte("{}")
	}

	buf := make([]byte, serverHeaderLen+len(payload))
	buf[0] = frame.Opcode
	binary.LittleEndian.PutUint16(buf[1:3], frame.Seq)
	buf[3] = frame.Phase
	binary.LittleEndian.PutUint16(buf[4:6], frame.EntityID)
	buf[6] = frame.Progress
	buf[7] = frame.Flags
	binary.LittleEndian.PutUint16(buf[8:10], uint16(len(payload)))
	copy(buf[10:], payload)
	return buf, nil
}

func DecodeServerFrame(data []byte) (ServerFrame, error) {
	payload := stripTransportPrefix(data)
	if len(payload) < serverHeaderLen {
		return ServerFrame{}, errors.New("server frame is too short")
	}

	size := int(binary.LittleEndian.Uint16(payload[8:10]))
	if len(payload) < serverHeaderLen+size {
		return ServerFrame{}, errors.New("server frame payload is truncated")
	}

	return ServerFrame{
		Opcode:   payload[0],
		Seq:      binary.LittleEndian.Uint16(payload[1:3]),
		Phase:    payload[3],
		EntityID: binary.LittleEndian.Uint16(payload[4:6]),
		Progress: payload[6],
		Flags:    payload[7],
		Payload:  append([]byte(nil), payload[10:10+size]...),
	}, nil
}

func MarshalPayload(payload ServerPayload) json.RawMessage {
	buf, _ := json.Marshal(payload)
	return buf
}

func stripTransportPrefix(data []byte) []byte {
	if len(data) > 1 && data[0]&0x80 != 0 {
		return data[1:]
	}
	return data
}
