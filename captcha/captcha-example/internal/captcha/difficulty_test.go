package captcha

import (
	"fmt"
	"testing"
)

func TestDifficultyProfileThresholds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input      int32
		stage      int
		subtasks   int
		ruleShift  bool
		serverPush bool
	}{
		{input: 1, stage: 1, subtasks: 1},
		{input: 20, stage: 1, subtasks: 1},
		{input: 21, stage: 2, subtasks: 1},
		{input: 60, stage: 4, subtasks: 2, ruleShift: true},
		{input: 80, stage: 5, subtasks: 3, ruleShift: true, serverPush: true},
		{input: 100, stage: 6, subtasks: 4, ruleShift: true, serverPush: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%d", tc.input), func(t *testing.T) {
			profile := NewDifficultyProfile(tc.input)
			if profile.Stage != tc.stage {
				t.Fatalf("stage mismatch: got %d want %d", profile.Stage, tc.stage)
			}
			if profile.Subtasks != tc.subtasks {
				t.Fatalf("subtasks mismatch: got %d want %d", profile.Subtasks, tc.subtasks)
			}
			if profile.HasRuleShift != tc.ruleShift {
				t.Fatalf("rule shift mismatch: got %v want %v", profile.HasRuleShift, tc.ruleShift)
			}
			if profile.HasServerPush != tc.serverPush {
				t.Fatalf("server push mismatch: got %v want %v", profile.HasServerPush, tc.serverPush)
			}
		})
	}
}
