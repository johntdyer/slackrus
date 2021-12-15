// Package slackrus provides a Slack hook for the logrus loggin package.
package slackrus

import (
	"fmt"
	"sort"

	"github.com/johntdyer/slack-go"
	"github.com/sirupsen/logrus"
)

// Project version
const (
	VERISON = "0.0.3"
)

// Filter is a filter applied to log entries to filter out any messages which are too noisy.
// The filter should return true if the message should be included, and false if not.
type Filter func(entry *logrus.Entry) bool

// SlackrusHook is a logrus Hook for dispatching messages to the specified
// channel on Slack.
type SlackrusHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	// Filters are applied to messages to determine if any entry should not be send out.
	Filters        []Filter
	HookURL        string
	IconURL        string
	Channel        string
	IconEmoji      string
	Username       string
	Asynchronous   bool
	Extra          map[string]interface{}
	Disabled       bool
	// SortFields if set to true will sort Fields before sending them to slack. By default they
	// are sorted in alphabetical order. For finer grained control, SortPriorities can be used.
	SortFields bool
	// SortPriorities if set will modify the straight alphabetical sort used when SortFields is set.
	// It is a map of field keys to sort priority, causing keys with higher priorities to appear first.
	// Any field field keys that do not appear in SortPriorities will appear after all those that do
	// and be sorted in alphabetical order.
	SortPriorities map[string]int
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
	if sh.Disabled {
		return nil
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

	// If any of the filters rejects the message then we return early
	for _, filter := range sh.Filters {
		if !filter(e) {
			return nil
		}
	}

	msg := &slack.Message{
		Username:  sh.Username,
		Channel:   sh.Channel,
		IconEmoji: sh.IconEmoji,
		IconUrl:   sh.IconURL,
	}

	attach := msg.NewAttachment()

	newEntry := sh.newEntry(e)
	// If there are fields we need to render them at attachments
	if len(newEntry.Data) > 0 {

		// Add a header above field data
		attach.Text = "Message fields"

		for k, v := range newEntry.Data {
			slackField := &slack.Field{}

			slackField.Title = k
			slackField.Value = fmt.Sprint(v)
			// If the field is <= 20 then we'll set it to short
			if len(slackField.Value) <= 20 {
				slackField.Short = true
			}

			attach.AddField(slackField)
		}
		attach.Pretext = newEntry.Message
	} else {
		attach.Text = newEntry.Message
	}
	attach.Fallback = newEntry.Message
	attach.Color = color

	if sh.SortFields {
		sort.SliceStable(attach.Fields, func(i, j int) bool {
			iTitle, jTitle := attach.Fields[i].Title, attach.Fields[j].Title
			if sh.SortPriorities == nil {
				return iTitle < jTitle
			}
			iVal, iOK := sh.SortPriorities[iTitle]
			jVal, jOK := sh.SortPriorities[jTitle]

			if iOK && !jOK {
				return true
			}
			if !iOK && jOK {
				return false
			}
			if (!iOK && !jOK) || iVal == jVal {
				return iTitle < jTitle
			}
			return iVal > jVal
		})
	}

	c := slack.NewClient(sh.HookURL)

	if sh.Asynchronous {
		go c.SendMessage(msg)
		return nil
	}

	return c.SendMessage(msg)
}

func (sh *SlackrusHook) newEntry(entry *logrus.Entry) *logrus.Entry {
	data := map[string]interface{}{}

	for k, v := range sh.Extra {
		data[k] = v
	}
	for k, v := range entry.Data {
		data[k] = v
	}

	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	return newEntry
}
