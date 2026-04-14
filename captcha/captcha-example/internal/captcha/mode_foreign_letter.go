package captcha

import (
	"fmt"
	"time"
)

type foreignLetterMode struct{}

type foreignLetterState struct {
	Waves []foreignWave
	Index int
}

type foreignWave struct {
	Prompt   string
	Items    []foreignItem
	Targets  int
	PushUsed bool
}

type foreignItem struct {
	Key    string
	Glyph  string
	Target bool
	Active bool
}

func newForeignLetterMode() Mode { return foreignLetterMode{} }

func (foreignLetterMode) ID() string       { return "foreign-letter" }
func (foreignLetterMode) Title() string    { return "Не русская буква" }
func (foreignLetterMode) HTMLFile() string { return "foreign-letter.html" }

func (m foreignLetterMode) InitSession(session *SessionState, now time.Time) *ServerAction {
	totalTargets := 4 + session.Difficulty.Subtasks
	waveCounts := splitCount(totalTargets, 3)
	state := &foreignLetterState{Waves: make([]foreignWave, 0, len(waveCounts))}
	for idx, count := range waveCounts {
		if count == 0 {
			continue
		}
		state.Waves = append(state.Waves, m.buildWave(session, idx, count))
	}

	session.Total = totalTargets
	session.ModeState = state
	m.planPush(session, state, now)

	return m.Snapshot(session, ServerOpInit, "Отмечайте буквы, которых нет в русском алфавите.")
}

func (m foreignLetterMode) Snapshot(session *SessionState, opcode byte, message string) *ServerAction {
	state := session.ModeState.(*foreignLetterState)
	return buildAction(session, opcode, message, func() ViewModel {
		wave := state.Waves[state.Index]
		options := make([]OptionView, 0, len(wave.Items))
		for _, item := range wave.Items {
			options = append(options, OptionView{
				ID:      session.NextEntity("letter", item.Key),
				Label:   item.Glyph,
				Accent:  "#0f172a",
				Variant: "glyph",
				Muted:   !item.Active,
			})
		}

		return ViewModel{
			Mode:         m.ID(),
			Theme:        "foreign-letter",
			Title:        "Отметьте чужие буквы",
			Instruction:  wave.Prompt,
			ProgressText: fmt.Sprintf("Волна %d из %d", state.Index+1, len(state.Waves)),
			Status:       message,
			Badges:       badgeSet(session.Difficulty),
			Layout: LayoutModel{
				Type:    "letters",
				Columns: 4,
				Options: options,
			},
		}
	})
}

func (m foreignLetterMode) HandleEvent(session *SessionState, frame ClientFrame, now time.Time) ModeOutcome {
	if frame.Opcode != ClientOpTap {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Нужно нажимать на буквы из текущей волны.")}
	}

	state := session.ModeState.(*foreignLetterState)
	wave := &state.Waves[state.Index]
	token := session.ActiveEntities[frame.Subject]
	index := -1
	for idx, item := range wave.Items {
		if item.Key == token.Key {
			index = idx
			break
		}
	}
	if index < 0 || !wave.Items[index].Active {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Этот символ уже обработан. Выберите другой.")}
	}
	if !wave.Items[index].Target {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Эта буква относится к текущему правилу и не подходит.")}
	}

	wave.Items[index].Active = false
	wave.Targets--
	session.Completed++

	if wave.Targets > 0 {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Верно. В этой волне еще есть лишние буквы.")}
	}

	if state.Index == len(state.Waves)-1 {
		session.Done = true
		return ModeOutcome{
			Action:   completionAction(session, m.ID(), m.Title(), "foreign-letter", "Все лишние буквы отмечены. Проверяем результат."),
			Finalize: true,
		}
	}

	state.Index++
	m.planPush(session, state, now)
	return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Волна закрыта. Следующая уже загружена.")}
}

func (m foreignLetterMode) Tick(session *SessionState, now time.Time) ModeOutcome {
	state, _ := session.ModeState.(*foreignLetterState)
	if state == nil || !session.Difficulty.HasServerPush {
		return ModeOutcome{}
	}

	push, ok := session.PopDuePush(now, m.ID())
	if !ok || push.Step != state.Index {
		return ModeOutcome{}
	}

	wave := &state.Waves[state.Index]
	if wave.PushUsed {
		return ModeOutcome{}
	}
	wave.PushUsed = true

	item := foreignItem{
		Key:    fmt.Sprintf("late-%d", state.Index),
		Glyph:  foreignUppercase[session.RNG.Intn(len(foreignUppercase))],
		Target: true,
		Active: true,
	}
	wave.Items = append(wave.Items, item)
	wave.Targets++
	session.Total++

	return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Появился новый символ. Проверьте волну еще раз.")}
}

func (m foreignLetterMode) buildWave(session *SessionState, index, targetCount int) foreignWave {
	upperOnly := session.Difficulty.HasRuleShift && index >= 1
	prompt := "Нажмите на все буквы, которых нет в русском алфавите."
	if upperOnly {
		prompt = "Нажмите только на ЗАГЛАВНЫЕ буквы, которых нет в русском алфавите."
	}

	items := make([]foreignItem, 0, 6)
	for i := 0; i < targetCount; i++ {
		glyph := foreignLowercase[session.RNG.Intn(len(foreignLowercase))]
		if upperOnly || (session.Difficulty.HasSecondaryCue && i%2 == 0) {
			glyph = foreignUppercase[session.RNG.Intn(len(foreignUppercase))]
		}
		items = append(items, foreignItem{
			Key:    fmt.Sprintf("target-%d-%d", index, i),
			Glyph:  glyph,
			Target: true,
			Active: true,
		})
	}

	for len(items) < 6 {
		glyph := russianLowercase[session.RNG.Intn(len(russianLowercase))]
		if session.RNG.Intn(2) == 0 {
			glyph = russianUppercase[session.RNG.Intn(len(russianUppercase))]
		}
		if upperOnly && session.RNG.Intn(3) == 0 {
			glyph = foreignLowercase[session.RNG.Intn(len(foreignLowercase))]
		}
		items = append(items, foreignItem{
			Key:    fmt.Sprintf("noise-%d-%d", index, len(items)),
			Glyph:  glyph,
			Target: false,
			Active: true,
		})
	}

	for i := len(items) - 1; i > 0; i-- {
		j := session.RNG.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}

	return foreignWave{
		Prompt:  prompt,
		Items:   items,
		Targets: targetCount,
	}
}

func (m foreignLetterMode) planPush(session *SessionState, state *foreignLetterState, now time.Time) {
	if !session.Difficulty.HasServerPush || state.Index >= len(state.Waves) {
		return
	}
	if state.Index < 2 {
		session.AddPush(scheduleDelay(now, session.Difficulty, state.Index), m.ID(), state.Index, 0)
	}
}
