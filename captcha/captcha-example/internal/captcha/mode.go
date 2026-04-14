package captcha

import (
	"encoding/json"
	"fmt"
	"time"
)

type ModeOutcome struct {
	Action   *ServerAction
	Finalize bool
}

type Mode interface {
	ID() string
	Title() string
	HTMLFile() string
	InitSession(*SessionState, time.Time) *ServerAction
	Snapshot(*SessionState, byte, string) *ServerAction
	HandleEvent(*SessionState, ClientFrame, time.Time) ModeOutcome
	Tick(*SessionState, time.Time) ModeOutcome
}

func buildMode(challengeID string) (Mode, error) {
	switch challengeID {
	case "odd-grid":
		return newOddGridMode(), nil
	case "reality-swipe":
		return newRealitySwipeMode(), nil
	case "foreign-letter":
		return newForeignLetterMode(), nil
	case "two-baskets":
		return newTwoBasketsMode(), nil
	case "track-object":
		return newTrackObjectMode(), nil
	default:
		return nil, fmt.Errorf("unsupported challenge id %q", challengeID)
	}
}

type ViewModel struct {
	Mode         string       `json:"mode"`
	Theme        string       `json:"theme"`
	Title        string       `json:"title"`
	Instruction  string       `json:"instruction"`
	ProgressText string       `json:"progressText"`
	Status       string       `json:"status,omitempty"`
	Layout       LayoutModel  `json:"layout"`
	Badges       []string     `json:"badges,omitempty"`
}

type LayoutModel struct {
	Type         string        `json:"type"`
	Columns      int           `json:"columns,omitempty"`
	Options      []OptionView  `json:"options,omitempty"`
	Card         *CardView     `json:"card,omitempty"`
	Baskets      []BucketView  `json:"baskets,omitempty"`
	Slots        []SlotView    `json:"slots,omitempty"`
	Sequence     []SwapView    `json:"sequence,omitempty"`
	AllowAnswer  bool          `json:"allowAnswer,omitempty"`
	HintLeft     string        `json:"hintLeft,omitempty"`
	HintRight    string        `json:"hintRight,omitempty"`
	Target       string        `json:"target,omitempty"`
	TargetAccent string        `json:"targetAccent,omitempty"`
}

type OptionView struct {
	ID      uint16 `json:"id"`
	Label   string `json:"label"`
	Icon    string `json:"icon,omitempty"`
	Accent  string `json:"accent,omitempty"`
	Variant string `json:"variant,omitempty"`
	Muted   bool   `json:"muted,omitempty"`
}

type CardView struct {
	ID      uint16 `json:"id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Icon    string `json:"icon,omitempty"`
	Accent  string `json:"accent,omitempty"`
}

type BucketView struct {
	ID     uint16 `json:"id"`
	Label  string `json:"label"`
	Accent string `json:"accent,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

type SlotView struct {
	ID      uint16 `json:"id"`
	Label   string `json:"label"`
	Token   string `json:"token,omitempty"`
	Accent  string `json:"accent,omitempty"`
	Active  bool   `json:"active,omitempty"`
	Correct bool   `json:"correct,omitempty"`
}

type SwapView struct {
	From    int `json:"from"`
	To      int `json:"to"`
	DelayMS int `json:"delayMs,omitempty"`
}

func buildAction(session *SessionState, opcode byte, message string, builder func() ViewModel) *ServerAction {
	session.BeginFrame()
	rootEntityID := session.NextEntity("frame", fmt.Sprintf("frame-%d", session.NextFrameSeq))
	view := builder()
	session.CurrentView = view

	payload, _ := json.Marshal(ServerPayload{
		Message: message,
		View:    view,
	})

	return &ServerAction{
		Frame: ServerFrame{
			Opcode:   opcode,
			Seq:      session.NextFrameSeq,
			Phase:    session.Phase,
			EntityID: rootEntityID,
			Progress: session.ProgressPercent(),
			Payload:  payload,
		},
	}
}
