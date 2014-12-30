package main

import (
	logrus "github.com/Sirupsen/logrus"
	"github.com/johntdyer/slackrus"
	"os"
)

func main() {

	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetOutput(os.Stderr)

	logrus.SetLevel(logrus.DebugLevel)

	logrus.AddHook(&slackrus.SlackrusHook{
		HookUrl:        "https://hooks.slack.com/services/abc123/defghijklmnopqrstuvwxyz",
		AcceptedLevels: slackrus.LevelThreshold(logrus.DebugLevel),
		Channel:        "#slack-testing",
		IconEmoji:      ":ghost:",
		Username:       "foobot",
	})

	logrus.Warn("warn")
	logrus.Info("info")
	logrus.Debug("debug")
}
