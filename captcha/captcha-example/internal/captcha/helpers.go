package captcha

import (
	"fmt"
	"time"
)

func completionAction(session *SessionState, modeID, title, theme, message string) *ServerAction {
	return buildAction(session, ServerOpResult, message, func() ViewModel {
		return ViewModel{
			Mode:         modeID,
			Theme:        theme,
			Title:        title,
			Instruction:  "Ответ принят. Проверяем результат.",
			ProgressText: "Готово",
			Status:       message,
			Layout: LayoutModel{
				Type: "complete",
				Card: &CardView{
					ID:     session.NextEntity("complete", "complete"),
					Title:  "Почти готово",
					Body:   message,
					Accent: "#22c55e",
				},
			},
		}
	})
}

func chooseCategory(rng randomSource, skip ...string) categorySet {
	for {
		idx := rng.Intn(len(categoryCatalog))
		candidate := categoryCatalog[idx]
		if !contains(skip, candidate.Key) {
			return candidate
		}
	}
}

func chooseSwipeScene(rng randomSource, possible bool) swipeScene {
	source := impossibleScenes
	if possible {
		source = possibleScenes
	}
	return source[rng.Intn(len(source))]
}

func chooseBasketRule(rng randomSource, skip ...string) basketRule {
	if len(skip) >= len(basketRules) {
		return basketRules[rng.Intn(len(basketRules))]
	}
	for {
		idx := rng.Intn(len(basketRules))
		candidate := basketRules[idx]
		if !contains(skip, candidate.Key) {
			return candidate
		}
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func pickItems(rng randomSource, items []visualItem, count int) []visualItem {
	indexes := rng.Perm(len(items))
	out := make([]visualItem, 0, count)
	for _, idx := range indexes {
		out = append(out, items[idx])
		if len(out) == count {
			break
		}
	}
	return out
}

func pickBasketItems(rng randomSource, items []basketItem, count int) []basketItem {
	indexes := rng.Perm(len(items))
	out := make([]basketItem, 0, count)
	for _, idx := range indexes {
		out = append(out, items[idx])
		if len(out) == count {
			break
		}
	}
	return out
}

func shuffleVisualItems(rng randomSource, items []visualItem) {
	for i := len(items) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}

func splitCount(total, segments int) []int {
	if segments <= 0 {
		return []int{total}
	}
	base := total / segments
	rest := total % segments
	out := make([]int, segments)
	for i := range out {
		out[i] = base
		if rest > 0 {
			out[i]++
			rest--
		}
	}
	return out
}

func scheduleDelay(now time.Time, profile DifficultyProfile, step int) time.Time {
	base := 900 + (step%3)*220
	if profile.HasServerPush {
		base -= 180
	}
	return now.Add(time.Duration(base) * time.Millisecond)
}

func badgeSet(profile DifficultyProfile) []string {
	out := []string{fmt.Sprintf("Сложность %d", profile.Normalized)}
	if profile.HasRuleShift {
		out = append(out, "Смена правила")
	}
	if profile.HasServerPush {
		out = append(out, "Живое обновление")
	}
	return out
}

type randomSource interface {
	Intn(int) int
	Perm(int) []int
}
