package captcha

import (
	"fmt"
	"testing"
	"time"
)

func TestModesCanReachCompletion(t *testing.T) {
	t.Parallel()

	modeIDs := []string{"odd-grid", "reality-swipe", "foreign-letter", "two-baskets", "track-object"}
	difficulties := []int32{20, 60, 80, 100}

	for _, modeID := range modeIDs {
		modeID := modeID
		for _, complexity := range difficulties {
			complexity := complexity
			t.Run(fmt.Sprintf("%s-%d", modeID, complexity), func(t *testing.T) {
				mode, err := buildMode(modeID)
				if err != nil {
					t.Fatalf("build mode: %v", err)
				}

				session := NewSessionState("session-"+modeID, mode.ID(), NewDifficultyProfile(complexity), time.Now())
				if action := mode.InitSession(session, time.Now()); action == nil {
					t.Fatalf("expected init action for %s", modeID)
				}

				for step := 0; step < 64 && !session.Done; step++ {
					switch state := session.ModeState.(type) {
					case *oddGridState:
						subject := lookupTokenID(session, state.Rounds[state.Index].Correct)
						outcome := mode.HandleEvent(session, ClientFrame{Opcode: ClientOpTap, Subject: subject}, time.Now())
						if outcome.Action == nil {
							t.Fatalf("odd-grid: expected response action")
						}
					case *realitySwipeState:
						dir := state.Rounds[state.Index].Correct
						outcome := mode.HandleEvent(session, ClientFrame{Opcode: ClientOpSwipe, Value: dir}, time.Now())
						if outcome.Action == nil {
							t.Fatalf("reality-swipe: expected response action")
						}
					case *foreignLetterState:
						key := ""
						for _, item := range state.Waves[state.Index].Items {
							if item.Target && item.Active {
								key = item.Key
								break
							}
						}
						if key == "" {
							t.Fatalf("foreign-letter: no active target left")
						}
						subject := lookupTokenID(session, key)
						outcome := mode.HandleEvent(session, ClientFrame{Opcode: ClientOpTap, Subject: subject}, time.Now())
						if outcome.Action == nil {
							t.Fatalf("foreign-letter: expected response action")
						}
					case *twoBasketsState:
						segment := state.Segments[state.Segment]
						card := segment.Cards[state.Card]
						subject := lookupTokenID(session, card.Key)
						target := lookupTokenID(session, "right")
						if card.LeftSide {
							target = lookupTokenID(session, "left")
						}
						outcome := mode.HandleEvent(session, ClientFrame{Opcode: ClientOpDragDrop, Subject: subject, Target: target}, time.Now())
						if outcome.Action == nil {
							t.Fatalf("two-baskets: expected response action")
						}
					case *trackObjectState:
						round := &state.Rounds[state.Index]
						if !round.AnswerOpen {
							for idx := range session.PendingPushes {
								session.PendingPushes[idx].At = time.Now().Add(-time.Second)
							}
							outcome := mode.Tick(session, time.Now().Add(2*time.Second))
							if outcome.Action == nil {
								t.Fatalf("track-object: expected tick action before answer")
							}
							continue
						}
						var subject uint16
						for idx, slot := range round.Slots {
							if slot.Key == round.TargetKey {
								subject = lookupTokenID(session, fmt.Sprintf("slot-%d", idx))
								break
							}
						}
						if subject == 0 {
							t.Fatalf("track-object: no target slot token")
						}
						outcome := mode.HandleEvent(session, ClientFrame{Opcode: ClientOpTap, Subject: subject}, time.Now())
						if outcome.Action == nil {
							t.Fatalf("track-object: expected response action")
						}
					default:
						t.Fatalf("unsupported mode state %T", state)
					}

					for {
						outcome := mode.Tick(session, time.Now().Add(2*time.Second))
						if outcome.Action == nil {
							break
						}
					}
				}

				if !session.Done {
					t.Fatalf("%s did not finish", modeID)
				}
			})
		}
	}
}

func lookupTokenID(session *SessionState, key string) uint16 {
	for id, token := range session.ActiveEntities {
		if token.Key == key {
			return id
		}
	}
	return 0
}
