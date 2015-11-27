// Package slackrus provides a Hipchat hook for the logrus loggin package.
package slackrus

import (
	"github.com/Sirupsen/logrus"
	"github.com/johntdyer/slack-go"
)

// Project version
const (
	VERISON = "0.0.2"
)

// SlackrusHook is a logrus Hook for dispatching messages to the specified
// channel on Slack.
type SlackrusHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	HookURL        string
	IconURL        string
	Channel        string
	IconEmoji      string
	Username       string
}

// Levels sets which levels to sent to slack
func (sh *SlackrusHook) Levels() []logrus.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

// Fire -  Sent event to slack
func (sh *SlackrusHook) Fire(e *logrus.Entry) error {
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
		Username:  sh.Username,
		Channel:   sh.Channel,
		IconEmoji: sh.IconEmoji,
		IconUrl:   sh.IconURL,
	}

	attach := msg.NewAttachment()

	// If there are fields we need to render them at attachments
	if len(e.Data) > 0 {

		// Add a header above field data
		attach.Text = "Message fields"

		for k, v := range e.Data {
			slackField := &slack.Field{}

			if str, ok := v.(string); ok {
				slackField.Title = k
				slackField.Value = str
				// If the field is <= 20 then we'll set it to short
				if len(str) <= 20 {
					slackField.Short = true
				}
			}
			attach.AddField(slackField)

		}
		attach.Pretext = e.Message
	} else {
		attach.Text = e.Message
	}
	attach.Fallback = e.Message
	attach.Color = color

	c := slack.NewClient(sh.HookURL)
	return c.SendMessage(msg)
}
