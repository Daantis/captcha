package captcha

import (
	"fmt"
	"time"
)

type realitySwipeMode struct{}

type realitySwipeState struct {
	Rounds []realityRound
	Index  int
}

type realityRound struct {
	Scene    swipeScene
	Correct  int16
	PushUsed bool
	Spare    *swipeScene
}

func newRealitySwipeMode() Mode { return realitySwipeMode{} }

func (realitySwipeMode) ID() string       { return "reality-swipe" }
func (realitySwipeMode) Title() string    { return "Что возможно?" }
func (realitySwipeMode) HTMLFile() string { return "reality-swipe.html" }

func (m realitySwipeMode) InitSession(session *SessionState, now time.Time) *ServerAction {
	total := 4 + session.Difficulty.Subtasks
	state := &realitySwipeState{Rounds: make([]realityRound, 0, total)}
	for i := 0; i < total; i++ {
		state.Rounds = append(state.Rounds, m.buildRound(session, i, total))
	}

	session.Total = total
	session.ModeState = state
	m.planPush(session, state, now)

	return m.Snapshot(session, ServerOpInit, "Свайпайте влево для невозможного, вправо для реального.")
}

func (m realitySwipeMode) Snapshot(session *SessionState, opcode byte, message string) *ServerAction {
	state := session.ModeState.(*realitySwipeState)
	return buildAction(session, opcode, message, func() ViewModel {
		round := state.Rounds[state.Index]
		return ViewModel{
			Mode:         m.ID(),
			Theme:        "reality-swipe",
			Title:        "Реально или невозможно",
			Instruction:  "Свайпайте карточку или используйте кнопки под ней.",
			ProgressText: fmt.Sprintf("Сцена %d из %d", state.Index+1, len(state.Rounds)),
			Status:       message,
			Badges:       badgeSet(session.Difficulty),
			Layout: LayoutModel{
				Type: "reality",
				Card: &CardView{
					ID:     session.NextEntity("scene", fmt.Sprintf("scene-%d", state.Index)),
					Title:  round.Scene.Title,
					Body:   round.Scene.Body,
					Icon:   round.Scene.Icon,
					Accent: round.Scene.Accent,
				},
				HintLeft:  "Невозможно",
				HintRight: "Возможно",
			},
		}
	})
}

func (m realitySwipeMode) HandleEvent(session *SessionState, frame ClientFrame, now time.Time) ModeOutcome {
	if frame.Opcode != ClientOpSwipe {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Здесь нужно свайпнуть карточку влево или вправо.")}
	}

	state := session.ModeState.(*realitySwipeState)
	round := &state.Rounds[state.Index]
	if frame.Value != round.Correct {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Направление не подходит. Посмотрите на сцену еще раз.")}
	}

	session.Completed++
	if state.Index == len(state.Rounds)-1 {
		session.Done = true
		return ModeOutcome{
			Action:   completionAction(session, m.ID(), m.Title(), "reality-swipe", "Карточки классифицированы. Проверяем результат."),
			Finalize: true,
		}
	}

	state.Index++
	m.planPush(session, state, now)
	return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Верно. Следующая карточка уже готова.")}
}

func (m realitySwipeMode) Tick(session *SessionState, now time.Time) ModeOutcome {
	state, _ := session.ModeState.(*realitySwipeState)
	if state == nil || !session.Difficulty.HasServerPush {
		return ModeOutcome{}
	}

	push, ok := session.PopDuePush(now, m.ID())
	if !ok || push.Step != state.Index {
		return ModeOutcome{}
	}

	round := &state.Rounds[state.Index]
	if round.PushUsed || round.Spare == nil {
		return ModeOutcome{}
	}

	round.PushUsed = true
	round.Scene = *round.Spare
	round.Correct = directionForScene(round.Scene)

	return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Карточка обновилась. Оцените новое условие.")}
}

func (m realitySwipeMode) buildRound(session *SessionState, index int, total int) realityRound {
	possible := session.RNG.Intn(2) == 0
	scene := chooseSwipeScene(session.RNG, possible)
	spare := chooseSwipeScene(session.RNG, !possible)

	if session.Difficulty.HasRuleShift && index >= total/2 {
		primary := chooseSwipeScene(session.RNG, session.RNG.Intn(2) == 0)
		focus := chooseSwipeScene(session.RNG, session.RNG.Intn(2) == 0)
		scene = swipeScene{
			Title:    "Смотрите только на вторую строку",
			Body:     primary.Body + "\n" + focus.Body,
			Icon:     focus.Icon,
			Accent:   focus.Accent,
			Possible: focus.Possible,
		}
		alt := chooseSwipeScene(session.RNG, !focus.Possible)
		spare = swipeScene{
			Title:    "Смотрите только на вторую строку",
			Body:     primary.Body + "\n" + alt.Body,
			Icon:     alt.Icon,
			Accent:   alt.Accent,
			Possible: alt.Possible,
		}
	}

	return realityRound{
		Scene:   scene,
		Correct: directionForScene(scene),
		Spare:   &spare,
	}
}

func (m realitySwipeMode) planPush(session *SessionState, state *realitySwipeState, now time.Time) {
	if !session.Difficulty.HasServerPush || state.Index >= len(state.Rounds) {
		return
	}
	if state.Index%2 == 1 {
		session.AddPush(scheduleDelay(now, session.Difficulty, state.Index), m.ID(), state.Index, 0)
	}
}

func directionForScene(scene swipeScene) int16 {
	if scene.Possible {
		return 1
	}
	return -1
}
