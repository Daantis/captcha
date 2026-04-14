package server

type captcherBuilder func(ChallengeId, InstanceId) (Captcher, error)

type ChallengeId string

func (t ChallengeId) String() string {
	return string(t)
}

type InstanceId string

func (i InstanceId) String() string {
	return string(i)
}
