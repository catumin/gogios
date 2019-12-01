package notifiers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bkasin/gogios/helpers"
)

// TelegramMessage uses a Telegram bot to post a notification that a check has changed states to a Telegram channel
func TelegramMessage(bot, channel, check, time, output string, status bool) error {
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

	urlString := "https://api.telegram.org/bot" + bot + "/sendMessage?chat_id=" + channel + "&text=" + message

	resp, err := http.Get(urlString)
	if err != nil {
		helpers.Log.Println("Error posting Telegram message, error: ", err)

		logger := helpers.AppendStringToFile("/var/log/gingertechnology/service_check.log", time+" Telegram post failed | "+resp.Status)
		if logger != nil {
			fmt.Println("Log could not be written. Error return:")
			fmt.Println(logger.Error())
		}
		return errors.New("failed to post Telegram message")
	}

	return nil
}
