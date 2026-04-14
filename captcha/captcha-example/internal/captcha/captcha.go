package captcha

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"sdk/event"
	"sdk/server"

	"github.com/google/uuid"
)

//go:embed all:html/*.html
var frontFS embed.FS

const (
	defaultSessionTTL = 3 * time.Minute
	tickInterval      = 250 * time.Millisecond
)

type Captcha struct {
	mode     Mode
	store    *SessionStore
	htmlByID map[string]string
	sendMu   sync.Mutex
}

func NewCaptcha(challengeType server.ChallengeId) (*Captcha, error) {
	mode, err := buildMode(challengeType.String())
	if err != nil {
		return nil, err
	}

	htmlByID, err := loadHTMLTemplates()
	if err != nil {
		return nil, err
	}

	return &Captcha{
		mode:     mode,
		store:    NewSessionStore(defaultSessionTTL),
		htmlByID: htmlByID,
	}, nil
}

func (c *Captcha) NewChallenge(_ context.Context, complexity int32) (*server.Challenge, error) {
	profile := NewDifficultyProfile(complexity)
	id := uuid.NewString()
	now := time.Now()

	session := NewSessionState(id, c.mode.ID(), profile, now)
	c.store.Put(session)

	html, ok := c.htmlByID[c.mode.HTMLFile()]
	if !ok {
		return nil, fmt.Errorf("html template %q is not embedded", c.mode.HTMLFile())
	}

	html = strings.ReplaceAll(html, "__CAPTCHA_ID__", id)
	html = strings.ReplaceAll(html, "__CAPTCHA_MODE__", c.mode.ID())
	html = strings.ReplaceAll(html, "__CAPTCHA_TITLE__", c.mode.Title())

	return &server.Challenge{
		Id:   id,
		HTML: html,
	}, nil
}

func (c *Captcha) HandleFrontendEvent(_ context.Context, stream server.Stream, challengeID string, data []byte) error {
	session, ok := c.store.Get(challengeID)
	if !ok {
		return nil
	}

	frame, err := DecodeClientEvent(data)
	if err != nil {
		slog.Warn("invalid client frame", "challenge_id", challengeID, "error", err.Error())
		session.mu.Lock()
		session.Errors++
		action := c.mode.Snapshot(session, ServerOpPrompt, "Соединение обновилось. Начните действие заново.")
		session.mu.Unlock()
		if action != nil {
			return c.sendAction(stream, challengeID, action)
		}
		return nil
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.Touch(time.Now())

	if frame.Opcode == ClientOpAckFrame {
		session.LastAckFrame = frame.ValueU16()
		return nil
	}

	if frame.Opcode == ClientOpReady {
		if session.Started {
			action := c.mode.Snapshot(session, ServerOpPrompt, "Задание уже открыто. Продолжайте с текущего шага.")
			if action != nil {
				return c.sendAction(stream, challengeID, action)
			}
			return nil
		}

		session.Started = true
		action := c.mode.InitSession(session, time.Now())
		if action != nil {
			return c.sendAction(stream, challengeID, action)
		}
		return nil
	}

	if !session.Started {
		action := c.mode.Snapshot(session, ServerOpPrompt, "Сначала дождитесь загрузки задания.")
		if action != nil {
			return c.sendAction(stream, challengeID, action)
		}
		return nil
	}

	if frame.Seq <= session.ClientSeq {
		session.Anomalies++
		action := c.mode.Snapshot(session, ServerOpPrompt, "Получена старая команда. Повторите действие.")
		if action != nil {
			return c.sendAction(stream, challengeID, action)
		}
		return nil
	}
	session.ClientSeq = frame.Seq

	if frame.Phase != session.Phase {
		session.Anomalies++
		action := c.mode.Snapshot(session, ServerOpPrompt, "Состояние успело измениться. Используйте обновленную версию задания.")
		if action != nil {
			return c.sendAction(stream, challengeID, action)
		}
		return nil
	}

	if frame.Opcode != ClientOpSwipe && frame.Opcode != ClientOpReady && frame.Opcode != ClientOpAckFrame {
		if frame.Subject == 0 {
			session.Errors++
			action := c.mode.Snapshot(session, ServerOpPrompt, "Нужно выбрать элемент из текущего задания.")
			if action != nil {
				return c.sendAction(stream, challengeID, action)
			}
			return nil
		}
		if _, ok := session.ActiveEntities[frame.Subject]; !ok {
			session.Anomalies++
			action := c.mode.Snapshot(session, ServerOpPrompt, "Элемент уже обновился. Повторите действие по новой версии задания.")
			if action != nil {
				return c.sendAction(stream, challengeID, action)
			}
			return nil
		}
	}

	if frame.Opcode == ClientOpDragDrop && frame.Target != 0 {
		if _, ok := session.ActiveEntities[frame.Target]; !ok {
			session.Anomalies++
			action := c.mode.Snapshot(session, ServerOpPrompt, "Корзина обновилась. Перетащите карточку еще раз.")
			if action != nil {
				return c.sendAction(stream, challengeID, action)
			}
			return nil
		}
	}

	session.UserEvents++
	outcome := c.mode.HandleEvent(session, frame, time.Now())
	if outcome.Action != nil {
		if err := c.sendAction(stream, challengeID, outcome.Action); err != nil {
			return err
		}
	}
	if outcome.Finalize {
		confidence := session.Confidence()
		c.store.Delete(challengeID)
		return c.sendResult(stream, challengeID, confidence)
	}

	return nil
}

func (c *Captcha) HandleConnectionClosed(_ context.Context, challengeID string, _ []byte) error {
	c.store.Delete(challengeID)
	return nil
}

func (c *Captcha) HandleBalancerEvent(_ context.Context, _ server.Stream, _ string, _ []byte) error {
	return nil
}

func (c *Captcha) HandleCloseChallenge(_ context.Context, _ server.Stream, challengeID string, _ []byte) error {
	c.store.Delete(challengeID)
	return nil
}

func (c *Captcha) OnStreamStarted(_ context.Context, _ server.Stream) error {
	return nil
}

func (c *Captcha) OnStreamClosed(_ context.Context, _ server.Stream) {}

func (c *Captcha) Stream(ctx context.Context, stream server.Stream) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.tick(stream, time.Now()); err != nil && !errors.Is(err, context.Canceled) {
				slog.Warn("tick send failed", "error", err.Error())
			}
		case <-ctx.Done():
			return
		case <-stream.Context().Done():
			return
		}
	}
}

func (c *Captcha) tick(stream server.Stream, now time.Time) error {
	for _, session := range c.store.Snapshot() {
		session.mu.Lock()

		if session.IsExpired(now) {
			session.mu.Unlock()
			c.store.Delete(session.ID)
			continue
		}

		if !session.Started || session.Done {
			session.mu.Unlock()
			continue
		}

		outcome := c.mode.Tick(session, now)
		session.mu.Unlock()

		if outcome.Action != nil {
			if err := c.sendAction(stream, session.ID, outcome.Action); err != nil {
				return err
			}
		}
		if outcome.Finalize {
			confidence := session.Confidence()
			c.store.Delete(session.ID)
			if err := c.sendResult(stream, session.ID, confidence); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Captcha) sendAction(stream server.Stream, challengeID string, action *ServerAction) error {
	if action == nil {
		return nil
	}

	data, err := EncodeServerFrame(action.Frame)
	if err != nil {
		return err
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	return stream.SendClientData(challengeID, &event.ServerEventClientData{Data: data})
}

func (c *Captcha) sendResult(stream server.Stream, challengeID string, confidence int32) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	return stream.SendChallengeResult(challengeID, &event.ServerEventChallengeResult{
		ConfidencePercent: confidence,
	})
}

func loadHTMLTemplates() (map[string]string, error) {
	files := []string{
		"odd-grid.html",
		"reality-swipe.html",
		"foreign-letter.html",
		"two-baskets.html",
		"track-object.html",
	}

	htmlByID := make(map[string]string, len(files))
	for _, name := range files {
		buf, err := frontFS.ReadFile("html/" + name)
		if err != nil {
			return nil, err
		}
		htmlByID[name] = string(buf)
	}

	return htmlByID, nil
}
