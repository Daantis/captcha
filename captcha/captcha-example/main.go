package main

import (
	"example/internal/captcha"
	"sdk/server"
)

func main() {
	server.SetCaptcherBuilder(func(challengeType server.ChallengeId, id server.InstanceId) (server.Captcher, error) {
		return captcha.NewCaptcha(challengeType)
	})
	server.Run()
}
