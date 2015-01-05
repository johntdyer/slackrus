// Package hiprus provides a Hipchat hook for the logrus loggin package.
package slackrus

import (
	"github.com/Sirupsen/logrus"
	"github.com/johntdyer/slack-go"
)

const (
	VERISON = "0.0.1"
)

var (
	client *slack.Client
)

// SlackrusHook is a logrus Hook for dispatching messages to the specified
// channel on Slack.
type SlackrusHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	HookUrl        string
	IconUrl        string
	Channel        string
	IconEmoji      string

	Username string
	c        *slack.Client
}

func (sh *SlackrusHook) Levels() []logrus.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

func (sh *SlackrusHook) Fire(e *logrus.Entry) error {
	if sh.c == nil {
		if err := sh.initClient(); err != nil {
			return err
		}
	}

	color := ""
	switch e.Level {
	case logrus.DebugLevel:
		color = "#9B30FF"
	case logrus.InfoLevel:
		color = "good"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}

	msg := &slack.Message{
		Username: sh.Username,
		Channel:  sh.Channel,
	}

	msg.IconEmoji = sh.IconEmoji
	msg.IconUrl = sh.IconUrl

	fmt.Println(e)

	attach := msg.NewAttachment()
	attach.Text = e.Fields //.Message
	attach.Color = color
	attach.Fallback = e.Fields //e.Message
	return sh.c.SendMessage(msg)

}

func (sh *SlackrusHook) initClient() error {
	sh.c = &slack.Client{sh.HookUrl}

	if sh.Username == "" {
		sh.Username = "SlackRus"
	}

	return nil
}
