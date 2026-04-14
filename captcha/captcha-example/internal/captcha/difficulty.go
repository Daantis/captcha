package captcha

type DifficultyProfile struct {
	Raw                 int32
	Normalized          int
	Stage               int
	Subtasks            int
	HasSecondaryCue     bool
	HasRuleShift        bool
	HasMultipleSegments bool
	HasServerPush       bool
}

func NewDifficultyProfile(raw int32) DifficultyProfile {
	value := int(raw)
	if value < 1 {
		value = 1
	}
	if value > 100 {
		value = 100
	}

	profile := DifficultyProfile{
		Raw:        raw,
		Normalized: value,
		Stage:      1,
		Subtasks:   1,
	}

	switch {
	case value >= 100:
		profile.Stage = 6
		profile.Subtasks = 4
	case value >= 80:
		profile.Stage = 5
		profile.Subtasks = 3
	case value >= 60:
		profile.Stage = 4
		profile.Subtasks = 2
	case value >= 41:
		profile.Stage = 3
	case value >= 21:
		profile.Stage = 2
	}

	profile.HasSecondaryCue = value >= 21
	profile.HasRuleShift = value >= 41
	profile.HasMultipleSegments = value >= 60
	profile.HasServerPush = value >= 80

	return profile
}
