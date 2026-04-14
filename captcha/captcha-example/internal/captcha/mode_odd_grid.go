package captcha

import (
	"fmt"
	"time"
)

type oddGridMode struct{}

type oddGridState struct {
	Rounds []oddGridRound
	Index  int
}

type oddGridRound struct {
	Prompt   string
	Items    []visualItem
	Correct  string
	Spare    *visualItem
	PushUsed bool
}

func newOddGridMode() Mode { return oddGridMode{} }

func (oddGridMode) ID() string       { return "odd-grid" }
func (oddGridMode) Title() string    { return "Выберите лишнее" }
func (oddGridMode) HTMLFile() string { return "odd-grid.html" }

func (m oddGridMode) InitSession(session *SessionState, now time.Time) *ServerAction {
	totalRounds := 4 + session.Difficulty.Subtasks
	state := &oddGridState{
		Rounds: make([]oddGridRound, 0, totalRounds),
	}

	for i := 0; i < totalRounds; i++ {
		state.Rounds = append(state.Rounds, m.buildRound(session, i, totalRounds))
	}

	session.Total = totalRounds
	session.ModeState = state
	m.planPush(session, state, now)

	return m.Snapshot(session, ServerOpInit, "Задание готово. Выберите один элемент.")
}

func (m oddGridMode) Snapshot(session *SessionState, opcode byte, message string) *ServerAction {
	state, _ := session.ModeState.(*oddGridState)
	if state == nil {
		return nil
	}

	return buildAction(session, opcode, message, func() ViewModel {
		round := state.Rounds[state.Index]
		options := make([]OptionView, 0, len(round.Items))
		for _, item := range round.Items {
			options = append(options, OptionView{
				ID:     session.NextEntity("odd-option", item.Key),
				Label:  item.Label,
				Icon:   item.Icon,
				Accent: item.Accent,
			})
		}

		return ViewModel{
			Mode:         m.ID(),
			Theme:        "odd-grid",
			Title:        "Найдите лишнее",
			Instruction:  round.Prompt,
			ProgressText: fmt.Sprintf("Раунд %d из %d", state.Index+1, len(state.Rounds)),
			Status:       message,
			Badges:       badgeSet(session.Difficulty),
			Layout: LayoutModel{
				Type:    "grid",
				Columns: 2,
				Options: options,
			},
		}
	})
}

func (m oddGridMode) HandleEvent(session *SessionState, frame ClientFrame, now time.Time) ModeOutcome {
	if frame.Opcode != ClientOpTap {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Нужно нажать на один из вариантов.")}
	}

	state := session.ModeState.(*oddGridState)
	round := &state.Rounds[state.Index]
	token := session.ActiveEntities[frame.Subject]
	if token.Key != round.Correct {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Это не лишний элемент. Попробуйте другой.")}
	}

	session.Completed++
	if state.Index == len(state.Rounds)-1 {
		session.Done = true
		return ModeOutcome{
			Action:   completionAction(session, m.ID(), m.Title(), "odd-grid", "Серия пройдена. Отправляем ответ на сервер."),
			Finalize: true,
		}
	}

	state.Index++
	m.planPush(session, state, now)
	return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Верно. Переходим к следующему раунду.")}
}

func (m oddGridMode) Tick(session *SessionState, now time.Time) ModeOutcome {
	state, _ := session.ModeState.(*oddGridState)
	if state == nil || !session.Difficulty.HasServerPush {
		return ModeOutcome{}
	}

	push, ok := session.PopDuePush(now, m.ID())
	if !ok || push.Step != state.Index {
		return ModeOutcome{}
	}

	round := &state.Rounds[state.Index]
	if round.PushUsed {
		return ModeOutcome{}
	}
	round.PushUsed = true

	if round.Spare != nil {
		for idx := range round.Items {
			if round.Items[idx].Key == round.Correct {
				continue
			}
			round.Items[idx] = *round.Spare
			break
		}
	}
	shuffleVisualItems(session.RNG, round.Items)

	return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Условие обновилось. Выберите лишний элемент по новой раскладке.")}
}

func (m oddGridMode) buildRound(session *SessionState, index int, total int) oddGridRound {
	target := chooseCategory(session.RNG)
	other := chooseCategory(session.RNG, target.Key)

	if session.Difficulty.HasRuleShift && index >= total/2 {
		targetItem := pickItems(session.RNG, target.Items, 1)[0]
		distractors := pickItems(session.RNG, other.Items, 2)
		third := chooseCategory(session.RNG, target.Key, other.Key)
		distractors = append(distractors, pickItems(session.RNG, third.Items, 1)[0])

		items := append([]visualItem{targetItem}, distractors...)
		shuffleVisualItems(session.RNG, items)

		spare := pickItems(session.RNG, other.Items, 1)[0]
		return oddGridRound{
			Prompt:  fmt.Sprintf("Нажмите на то, что %s.", target.Prompt),
			Items:   items,
			Correct: targetItem.Key,
			Spare:   &spare,
		}
	}

	mainItems := pickItems(session.RNG, target.Items, 3)
	odd := pickItems(session.RNG, other.Items, 1)[0]
	items := append(mainItems, odd)
	shuffleVisualItems(session.RNG, items)
	spare := pickItems(session.RNG, other.Items, 1)[0]

	return oddGridRound{
		Prompt:  "Нажмите на предмет, который не подходит к остальным.",
		Items:   items,
		Correct: odd.Key,
		Spare:   &spare,
	}
}

func (m oddGridMode) planPush(session *SessionState, state *oddGridState, now time.Time) {
	if !session.Difficulty.HasServerPush || state.Index >= len(state.Rounds) {
		return
	}

	if state.Index%2 == 0 {
		session.AddPush(scheduleDelay(now, session.Difficulty, state.Index), m.ID(), state.Index, 0)
	}
}
