package captcha

import (
	"fmt"
	"time"
)

type twoBasketsMode struct{}

type twoBasketsState struct {
	Segments []basketSegment
	Segment  int
	Card     int
}

type basketSegment struct {
	Rule     basketRule
	Cards    []basketItem
	PushUsed bool
}

func newTwoBasketsMode() Mode { return twoBasketsMode{} }

func (twoBasketsMode) ID() string       { return "two-baskets" }
func (twoBasketsMode) Title() string    { return "Разложите по корзинам" }
func (twoBasketsMode) HTMLFile() string { return "two-baskets.html" }

func (m twoBasketsMode) InitSession(session *SessionState, now time.Time) *ServerAction {
	totalCards := 4 + session.Difficulty.Subtasks
	segments := session.Difficulty.Subtasks
	counts := splitCount(totalCards, segments)
	state := &twoBasketsState{Segments: make([]basketSegment, 0, len(counts))}

	var used []string
	for idx, count := range counts {
		rule := chooseBasketRule(session.RNG, used...)
		if !session.Difficulty.HasRuleShift && idx > 0 {
			rule = state.Segments[0].Rule
		}
		used = append(used, rule.Key)
		state.Segments = append(state.Segments, basketSegment{
			Rule:  rule,
			Cards: pickBasketItems(session.RNG, rule.Items, count),
		})
	}

	session.Total = totalCards
	session.ModeState = state
	m.planPush(session, state, now)

	return m.Snapshot(session, ServerOpInit, "Перетащите карточку в правильную корзину.")
}

func (m twoBasketsMode) Snapshot(session *SessionState, opcode byte, message string) *ServerAction {
	state := session.ModeState.(*twoBasketsState)
	return buildAction(session, opcode, message, func() ViewModel {
		segment := state.Segments[state.Segment]
		card := segment.Cards[state.Card]
		return ViewModel{
			Mode:         m.ID(),
			Theme:        "two-baskets",
			Title:        "Разложите по корзинам",
			Instruction:  fmt.Sprintf("Текущее правило: %s / %s.", segment.Rule.LeftLabel, segment.Rule.RightLabel),
			ProgressText: fmt.Sprintf("Карточка %d из %d", session.Completed+1, session.Total),
			Status:       message,
			Badges:       badgeSet(session.Difficulty),
			Layout: LayoutModel{
				Type: "baskets",
				Baskets: []BucketView{
					{
						ID:     session.NextEntity("basket", "left"),
						Label:  segment.Rule.LeftLabel,
						Accent: segment.Rule.LeftAccent,
						Hint:   "Левая корзина",
					},
					{
						ID:     session.NextEntity("basket", "right"),
						Label:  segment.Rule.RightLabel,
						Accent: segment.Rule.RightAccent,
						Hint:   "Правая корзина",
					},
				},
				Options: []OptionView{
					{
						ID:     session.NextEntity("card", card.Key),
						Label:  card.Label,
						Icon:   card.Icon,
						Accent: "#ffffff",
					},
				},
			},
		}
	})
}

func (m twoBasketsMode) HandleEvent(session *SessionState, frame ClientFrame, now time.Time) ModeOutcome {
	if frame.Opcode != ClientOpDragDrop {
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Здесь нужно перетащить карточку в корзину.")}
	}

	state := session.ModeState.(*twoBasketsState)
	segment := &state.Segments[state.Segment]
	card := segment.Cards[state.Card]
	cardToken := session.ActiveEntities[frame.Subject]
	basketToken := session.ActiveEntities[frame.Target]

	if cardToken.Key != card.Key {
		session.Anomalies++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Карточка уже обновилась. Перетащите новую.")}
	}

	correctBasket := "right"
	if card.LeftSide {
		correctBasket = "left"
	}
	if basketToken.Key != correctBasket {
		session.Errors++
		return ModeOutcome{Action: m.Snapshot(session, ServerOpPrompt, "Карточка должна попасть в другую корзину.")}
	}

	session.Completed++
	if state.Card == len(segment.Cards)-1 {
		if state.Segment == len(state.Segments)-1 {
			session.Done = true
			return ModeOutcome{
				Action:   completionAction(session, m.ID(), m.Title(), "two-baskets", "Карточки разложены. Проверяем результат."),
				Finalize: true,
			}
		}
		state.Segment++
		state.Card = 0
		m.planPush(session, state, now)
		return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Сегмент завершен. Загружаем новое правило.")}
	}

	state.Card++
	return ModeOutcome{Action: m.Snapshot(session, ServerOpProgress, "Верно. Следующая карточка уже на столе.")}
}

func (m twoBasketsMode) Tick(session *SessionState, now time.Time) ModeOutcome {
	state, _ := session.ModeState.(*twoBasketsState)
	if state == nil || !session.Difficulty.HasServerPush {
		return ModeOutcome{}
	}

	push, ok := session.PopDuePush(now, m.ID())
	if !ok || push.Step != state.Segment {
		return ModeOutcome{}
	}

	segment := &state.Segments[state.Segment]
	if segment.PushUsed {
		return ModeOutcome{}
	}
	segment.PushUsed = true

	spare := pickBasketItems(session.RNG, segment.Rule.Items, 1)[0]
	segment.Cards = append(segment.Cards, spare)
	session.Total++

	return ModeOutcome{Action: m.Snapshot(session, ServerOpPatch, "Появилась дополнительная карточка. Разложите и ее тоже.")}
}

func (m twoBasketsMode) planPush(session *SessionState, state *twoBasketsState, now time.Time) {
	if !session.Difficulty.HasServerPush || state.Segment >= len(state.Segments) {
		return
	}
	session.AddPush(scheduleDelay(now, session.Difficulty, state.Segment), m.ID(), state.Segment, 0)
}
