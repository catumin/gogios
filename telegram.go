package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// PostMessage uses a Telegram bot to post a notification that a check has failed to a Telegram channel
func PostMessage(bot, channel, check, time, output string, status bool) error {
	newlines := regexp.MustCompile(`\r?\n`)
	message := ""
	if status {
		message = newlines.ReplaceAllString(check+" Status changed to Success as of:\n"+time+"\n\nOutput of check was:\n"+output, "%0A")
	} else {
		message = newlines.ReplaceAllString(check+" Status changed to Fail as of:\n"+time+"\n\nOutput of check was:\n"+output, "%0A")
	}

	urlString := "https://api.telegram.org/bot" + bot + "/sendMessage?chat_id=" + channel + "&text=" + message

	resp, err := http.Get(urlString)
	if err != nil {
		fmt.Println("Error posting Telegram message, error: ", err)
		logerr := AppendStringToFile("/var/log/gingertechnology/service_check.log", time+" Telegram post failed | "+resp.Status)
		if logerr != nil {
			fmt.Println("Log could not be written. God save you, error return:")
			fmt.Println(logerr.Error())
		}
		return errors.New("failed to post Telegram message")
	}

	return nil
}
