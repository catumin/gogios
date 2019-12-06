package twilio

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/notifiers"
)

type Twilio struct {
	SID             string
	Token           string
	TwilioNumber    string `toml:"twilio_number"`
	SendTo          string `toml:"send_to"`
	ResponseTimeout helpers.Duration

	// HTTP Client
	client *http.Client
}

var sampleConfig = `
  ## Account SID
  sid = ""
  ## Account auth token
  token = ""
  ## Phone number for your Twilio account
  twilio_number = ""
  ## Phone number that should be texted with the notification
  send_to = ""

  ## HTTP response timeout (default: 10s)
  response_timeout = "10s"
`

func (t *Twilio) SampleConfig() string {
	return sampleConfig
}

func (t *Twilio) Description() string {
	return "Send a text message to a phone number using Twilio's REST API"
}

func (t *Twilio) Notify(check, time, output string, status bool) error {
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

	urlString := "https://api.twilio.com/2010-04-01/Accounts/" + t.SID + "/Messages.json"
	// Pack up the data for our message
	msgData := url.Values{}
	msgData.Set("To", t.SendTo)
	msgData.Set("From", t.TwilioNumber)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	req, _ := http.NewRequest("POST", urlString, &msgDataReader)
	req.SetBasicAuth(t.SID, t.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make HTTP POST request and return message SID
	resp, _ := t.client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			helpers.Log.Println("Twilio message posted. SID: ", data["sid"])
		}
	} else {
		helpers.Log.Println("Error sending Twilio message. Return was: ", resp.Status)
	}

	return nil
}

func (t *Twilio) createHTTPClient() (*http.Client, error) {
	if t.ResponseTimeout.Duration < time.Second {
		t.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Timeout: t.ResponseTimeout.Duration,
	}

	return client, nil
}

func init() {
	notifiers.Add("twilio", func() gogios.Notifier {
		return &Twilio{}
	})
}
