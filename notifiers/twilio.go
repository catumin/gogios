package notifiers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TwilioMessage uses the Twilio REST API to sned a text message alerting that a check has changed states
func TwilioMessage(sid, token, twilioNumber, sendTo, check, time, output string, status bool) error {
	message := ""
	tailedOutput := output

	if len(output) > 1250 {
		runeOutput := []rune(output)
		tailedOutput = string(runeOutput[len(output)-1250:])
	}
	if status {
		message = url.QueryEscape(check + " Status changed to Success as of:\n" + time + "\n\nOutput of check was:\n" + tailedOutput)
	} else {
		message = url.QueryEscape(check + " Status changed to Fail as of:\n" + time + "\n\nOutput of check was:\n" + tailedOutput)
	}

	urlString := "https://api.twilio.com/2010-04-01/Accounts/" + sid + "/Messages.json"
	// Pack up the data for our message
	msgData := url.Values{}
	msgData.Set("To", sendTo)
	msgData.Set("From", twilioNumber)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	// Create HTTP request client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlString, &msgDataReader)
	req.SetBasicAuth(sid, token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make HTTP POST request and return message SID
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}

	return nil
}
