package slackrus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johntdyer/slack-go"
	"github.com/sirupsen/logrus"
)

type Fixture struct {
	Messages []slack.Message
	Cleanup  func()
	MsgRcvd  chan struct{}

	server *httptest.Server
}

func (f *Fixture) URL() string {
	return f.server.URL
}

func NewFixture(t *testing.T) *Fixture {
	f := &Fixture{MsgRcvd: make(chan struct{}, 1)}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg slack.Message
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&msg)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		f.Messages = append(f.Messages, msg)
		select {
		case f.MsgRcvd <- struct{}{}:
		default:
		}
	}))
	f.Cleanup = func() { f.server.Close() }
	return f
}

func TestFieldSorting(t *testing.T) {
	f := NewFixture(t)
	defer f.Cleanup()

	logger := logrus.New()
	logger.AddHook(&SlackrusHook{
		HookURL:    f.URL(),
		Channel:    "#slack-testing",
		SortFields: true,
	})

	first, second, third := "a_should_be_first", "b_should_be_second", "c_should_be_third"

	logger.WithFields(logrus.Fields{
		second: "b content",
		third:  "c content",
		first:  "a content",
	}).Info("well hello there, you better sort my fields!!!!")

	<-f.MsgRcvd

	if exp, got := 1, len(f.Messages); exp != got {
		t.Fatalf("received unexpected number of messages: exp: %d, got: %d", exp, got)
	}
	msg := f.Messages[0]
	if exp, got := 1, len(msg.Attachments); exp != got {
		t.Fatalf("received unexpected number of Attachments in message: exp: %d, got: %d", exp, got)
	}
	fields := msg.Attachments[0].Fields
	if exp, got := 3, len(fields); exp != got {
		t.Fatalf("received unexpected number of Fields in attachment: exp: %d, got: %d", exp, got)
	}

	if exp, got := first, fields[0].Title; exp != got {
		t.Errorf("0-th field title not as expected: exp: %q, got: %q", exp, got)
	}
	if exp, got := second, fields[1].Title; exp != got {
		t.Errorf("1st field title not as expected: exp: %q, got: %q", exp, got)
	}
	if exp, got := third, fields[2].Title; exp != got {
		t.Errorf("2nd field title not as expected: exp: %q, got: %q", exp, got)
	}
}

func TestFieldSortPriorities(t *testing.T) {
	f := NewFixture(t)
	defer f.Cleanup()

	first, second, third, fourth := "d_should_be_first", "b_should_be_second", "a_should_be_third", "c_should_be_fourth"

	logger := logrus.New()
	logger.AddHook(&SlackrusHook{
		HookURL:    f.URL(),
		Channel:    "#slack-testing",
		SortFields: true,
		SortPriorities: map[string]int{
			first:  10,
			second: 5,
		},
	})

	logger.WithFields(logrus.Fields{
		first:  "first content",
		second: "second content",
		third:  "third content",
		fourth: "fourth content",
	}).Info("well hello there, you better sort my fields!!!!")

	<-f.MsgRcvd

	if exp, got := 1, len(f.Messages); exp != got {
		t.Fatalf("received unexpected number of messages: exp: %d, got: %d", exp, got)
	}
	msg := f.Messages[0]
	if exp, got := 1, len(msg.Attachments); exp != got {
		t.Fatalf("received unexpected number of Attachments in message: exp: %d, got: %d", exp, got)
	}
	fields := msg.Attachments[0].Fields
	if exp, got := 4, len(fields); exp != got {
		t.Fatalf("received unexpected number of Fields in attachment: exp: %d, got: %d", exp, got)
	}

	if exp, got := first, fields[0].Title; exp != got {
		t.Errorf("0-th field title not as expected: exp: %q, got: %q", exp, got)
	}
	if exp, got := second, fields[1].Title; exp != got {
		t.Errorf("1st field title not as expected: exp: %q, got: %q", exp, got)
	}
	if exp, got := third, fields[2].Title; exp != got {
		t.Errorf("2nd field title not as expected: exp: %q, got: %q", exp, got)
	}
	if exp, got := fourth, fields[3].Title; exp != got {
		t.Errorf("3rd field title not as expected: exp: %q, got: %q", exp, got)
	}
}
