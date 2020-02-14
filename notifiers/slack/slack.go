package slack

import (
	"fmt"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/notifiers"
	"github.com/nlopes/slack"
)

type Slack struct {
	Token           string
	Channel         string `toml:"channel"`
	ResponseTimeout helpers.Duration
}

var sampleConfig = `
  ## Slack bot API token
  token = ""
  ## Channel ID to post message to
  chats = ""
`

func (s *Slack) SampleConfig() string {
	return sampleConfig
}

func (s *Slack) Description() string {
	return "Send a notification to a Slack channel using a bot when a check changes states"
}

func (s *Slack) Notify(check, time, output, status string) error {
	api := slack.New(s.Token)
	attachment := slack.Attachment{
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Output",
				Value: output,
			},
		},
	}

	channelID, timestamp, err := api.PostMessage(s.Channel, slack.MsgOptionText(check+" Status changed to "+status+" as of:\n"+time+"\n\nOutput of check was:", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		return err
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

	return nil
}

func (s *Slack) Init() error {
	return nil
}

func init() {
	notifiers.Add("slack", func() gogios.Notifier {
		return &Slack{}
	})
}
