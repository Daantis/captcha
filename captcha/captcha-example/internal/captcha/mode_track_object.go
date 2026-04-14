package captcha

import (
	"fmt"
	"time"
)

type trackObjectMode struct{}

type trackObjectState struct {
	Rounds []trackRound
	Index  int
}

type trackRound struct {
	Slots        []trackPiece
	TargetKey    string
	TargetLabel  string
	TargetAccent string
	Swaps        []SwapView
	SwapIndex    int
	AnswerOpen   bool
	LateUsed     bool
	LastSwap     []SwapView
}

func newTrackObjectMode() Mode { return trackObjectMode{} }

func (trackObjectMode) ID() string       { return "track-object" }
func (trackObjectMode) Title() string    { return "Проследите за фишкой" }
func (trackObjectMode) HTMLFile() string { return "track-object.html" }

func (m trackObjectMode) InitSession(session *SessionState, now time.Time) *ServerAction {
	totalRounds := 4 + session.Difficulty.Subtasks
	state := &trackObjectState{Rounds: make([]trackRound, 0, totalRounds)}
	for i := 0; i < totalRounds; i++ {
		state.Rounds = append(state.Rounds, m.buildRound(session, i, totalRounds))
	}

	session.Total = totalRounds
	session.ModeState = state
	m.scheduleRound(session, state, now)

	return m.Snapshot(session, ServerOpInit, "Следите за фишкой. Сервер будет двигать ее сам.")
}

func (m trackObjectMode) Snapshot(session *SessionState, opcode byte, message string) *ServerAction {
	state := session.ModeState.(*trackObjectState)
	return buildAction(session, opcode, message, func() ViewModel {
		round := state.Rounds[state.Index]
		slots := make([]SlotView, 0, len(round.Slots))
		for idx, slot := range round.Slots {
			key := fmt.Sprintf("slot-%d", idx)
			slots = append(slots, SlotView{
				ID:     session.NextEntity("slot", key),
				Label:  fmt.Sprintf("%d", idx+1),
				Token:  slot.Label,
				Accent: slot.Accent,
				Active: round.AnswerOpen,
			})
		}

		instruction := "Запомните фишку и дождитесь конца движения."
		if round.AnswerOpen {
			instruction = "Теперь нажмите на окно, где остановилась нужная фишка."
		}

		return ViewModel{
			Mode:         m.ID(),
			Theme:        "track-object",
			Title:        "Проследите за фишкой",
			Instruction:  instruction,
			ProgressText: fmt.Sprintf("Раунд %d из %d", state.Index+1, len(state.Rounds)),
			Status:       message,
			Badges:       badgeSet(session.Difficulty),
			Layout: LayoutModel{
				Type:         "track",
				Slots:        slots,
				Sequence:     round.LastSwap,
				AllowAnswer:  round.AnswerOpen,
				Target:       round.TargetLabel,
				TargetAccent: round.TargetAccent,
			},
		}
	})
}

func (m trackObjectMode) HandleEvent(session *SessionState, frame ClientFrame, now time.Time) ModeOutcome {
	if frame.Opcode != ClientOpTap {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "В конце раунда нужно нажать на нужное окно.")}
	}

	state := session.ModeState.(*trackObjectState)
	round := &state.Rounds[state.Index]
	if !round.AnswerOpen {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Подождите завершения движения, затем выберите окно.")}
	}

	token := session.ActiveEntities[frame.Subject]
	slotIndex := -1
	for idx := range round.Slots {
		if token.Key == fmt.Sprintf("slot-%d", idx) {
			slotIndex = idx
			break
		}
	}
	if slotIndex < 0 {
		session.Anomalies++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Окно обновилось. Нажмите по новой версии доски.")}
	}

	if round.Slots[slotIndex].Key != round.TargetKey {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Это не та фишка. Посмотрите на позицию еще раз.")}
	}

	session.Completed++
	if state.Index == len(state.Rounds)-1 {
		session.Done = true
		return ModeOutcome{
			Action:   completionAction(session, m.ID(), m.Title(), "track-object", "Треки завершены. Проверяем результат."),
			Finalize: true,
		}
	}

	state.Index++
	m.scheduleRound(session, state, now)
	return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Верно. Начинаем следующий раунд.")}
}

func (m trackObjectMode) Tick(session *SessionState, now time.Time) ModeOutcome {
	state, _ := session.ModeState.(*trackObjectState)
	if state == nil {
		return ModeOutcome{}
	}

	push, ok := session.PopDuePush(now, m.ID())
	if !ok || push.Step != state.Index {
		return ModeOutcome{}
	}

	round := &state.Rounds[state.Index]
	if round.AnswerOpen {
		return ModeOutcome{}
	}

	if round.SwapIndex < len(round.Swaps) {
		swap := round.Swaps[round.SwapIndex]
		round.LastSwap = []SwapView{swap}
		m.applySwap(round, swap)
		round.SwapIndex++
		if round.SwapIndex < len(round.Swaps) {
			session.AddPush(now.Add(650*time.Millisecond), m.ID(), state.Index, 0)
		} else if session.Difficulty.HasServerPush && !round.LateUsed {
			round.LateUsed = true
			extra := SwapView{From: 1, To: len(round.Slots), DelayMS: 240}
			if len(round.Slots) > 3 {
				extra = SwapView{From: 2, To: 4, DelayMS: 240}
			}
			round.Swaps = append(round.Swaps, extra)
			session.AddPush(now.Add(650*time.Millisecond), m.ID(), state.Index, 0)
		} else {
			round.AnswerOpen = true
		}

		return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Сервер двинул фишки. Следите за доской.")}
	}

	round.AnswerOpen = true
	round.LastSwap = nil
	return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Движение закончено. Выберите нужное окно.")}
}

func (m trackObjectMode) buildRound(session *SessionState, index int, total int) trackRound {
	slots := append([]trackPiece(nil), trackPieces...)
	for i := len(slots) - 1; i > 0; i-- {
		j := session.RNG.Intn(i + 1)
		slots[i], slots[j] = slots[j], slots[i]
	}

	target := slots[session.RNG.Intn(len(slots))]
	if session.Difficulty.HasRuleShift && index >= total/2 {
		target = slots[1]
	}

	swapCount := 3
	if session.Difficulty.HasSecondaryCue {
		swapCount = 4
	}
	swaps := make([]SwapView, 0, swapCount)
	for i := 0; i < swapCount; i++ {
		from := session.RNG.Intn(len(slots)) + 1
		to := session.RNG.Intn(len(slots)) + 1
		for to == from {
			to = session.RNG.Intn(len(slots)) + 1
		}
		swaps = append(swaps, SwapView{From: from, To: to, DelayMS: 220})
	}

	label := target.Label
	if session.Difficulty.HasRuleShift && index >= total/2 {
		label = "Фишка, которая стартовала во втором окне"
	}

	return trackRound{
		Slots:        slots,
		TargetKey:    target.Key,
		TargetLabel:  label,
		TargetAccent: target.Accent,
		Swaps:        swaps,
	}
}

func (m trackObjectMode) applySwap(round *trackRound, swap SwapView) {
	from := swap.From - 1
	to := swap.To - 1
	if from < 0 || from >= len(round.Slots) || to < 0 || to >= len(round.Slots) {
		return
	}
	round.Slots[from], round.Slots[to] = round.Slots[to], round.Slots[from]
}

func (m trackObjectMode) scheduleRound(session *SessionState, state *trackObjectState, now time.Time) {
	if state.Index >= len(state.Rounds) {
		return
	}
	session.AddPush(now.Add(600*time.Millisecond), m.ID(), state.Index, 0)
}
