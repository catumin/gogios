package telegram

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/notifiers"
)

type Telegram struct {
	API             string
	Chats           []string `toml:"chats"`
	ResponseTimeout helpers.Duration

	// HTTP Client
	client *http.Client
}

var sampleConfig = `
  ## Telegram bot API key
  api = ""
  ## Chat to post to, usually a negative number for groups and a positive one for direct
  chats = [""]

  ## HTTP response timeout (default: 10s)
  response_timeout = "10s"
`

func (t *Telegram) SampleConfig() string {
	return sampleConfig
}

func (t *Telegram) Description() string {
	return "Send a notification to a Telegram channel using a bot when a check changes states"
}

func (t *Telegram) Notify(check, time, output string, status bool) error {
	var wg sync.WaitGroup

	// Shorten the message if it too long for a URL
	message := ""
	tailedOutput := output
	if len(output) > 1250 {
		runeOutput := []rune(output)
		tailedOutput = string(runeOutput[len(output)-1250:])
	}

	// Build the message
	if status {
		message = url.QueryEscape(check + " Status changed to Success as of:\n" + time + "\n\nOutput of check was:\n" + tailedOutput)
	} else {
		message = url.QueryEscape(check + " Status changed to Fail as of:\n" + time + "\n\nOutput of check was:\n" + tailedOutput)
	}

	// Create the HTTP Client
	if t.client == nil {
		client, err := t.createHTTPClient()
		if err != nil {
			return err
		}

		t.client = client
	}

	for _, c := range t.Chats {
		u := "https://api.telegram.org/bot" + t.API + "/sendMessage?chat_id=" + c + "&text=" + message
		addr, err := url.Parse(u)
		if err != nil {
			helpers.Log.Println(fmt.Errorf("Telegram: Unable to parse address:\n'%s'\n\n%s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()

			resp, err := t.client.Get(addr.String())
			if err != nil {
				helpers.Log.Println(err.Error)
				return
			}

			helpers.Log.Printf("Telegram message posted: %s", resp.Status)
		}(addr)
	}

	wg.Wait()
	return nil
}

func (t *Telegram) createHTTPClient() (*http.Client, error) {
	if t.ResponseTimeout.Duration < time.Second {
		t.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Timeout: t.ResponseTimeout.Duration,
	}

	return client, nil
}

func init() {
	notifiers.Add("telegram", func() gogios.Notifier {
		return &Telegram{}
	})
}
