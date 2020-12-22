//принимать токен сервиса переводчика
//откидывать протухшие сообщения
//отправлять и принимать сообщения в/от сервис-переводчик
//определять язык сообщения

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type resStruct struct {
	Err                   string
	Result                string
	CacheUse              int
	Source                string
	From                  string
	SourceTransliteration string
	TargetTransliteration string
}

type reqStruct struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Data     string `json:"data"`
	Platform string `json:"platform"`
}

func main() {

	var tokenTGBot string
	var tokenLgvnx string
	const tokenStringTg = "tgkey"
	const tokenStringLn = "lnkey"

	flag.StringVar(&tokenTGBot, tokenStringTg, "", "The token of your telegramm bot")
	flag.StringVar(&tokenLgvnx, tokenStringLn, "", "Your token to lingvanex API")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "About: Telegram bot for translating from english to russian and back\n")
		fmt.Fprintf(os.Stderr, "Author: Karmanov Mikhail\n")
		fmt.Fprintf(os.Stderr, "Version: 0.0.1\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if tokenTGBot == "" {
		log.Panic("Please insert token of your telegramm bot!")
	}

	bot, err := tgbotapi.NewBotAPI(tokenTGBot)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		textInMessage := update.Message.Text

		myText := sendToTranslater(tokenLgvnx, textInMessage)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, myText)

		bot.Send(msg)
	}
}

func sendToTranslater(apiKey string, sendingText string) string {
	const codeEn = "en_GB"
	const codeRu = "ru_RU"
	const myPlatform = "api"

	const url = "https://api-b2b.backenster.com/b1/api/v3/translate"

	var codeFrom string
	var codeTo string

	codeFrom = codeEn
	codeTo = codeRu
	if isCyrillic(sendingText) {
		codeFrom = codeRu
		codeTo = codeEn
	}

	var requestJson reqStruct
	requestJson.From = codeFrom
	requestJson.To = codeTo
	requestJson.Data = sendingText
	requestJson.Platform = myPlatform

	jsonSendData, err0 := json.Marshal(requestJson)
	if err0 != nil {
		panic(err0)
		return "fault to translate"
	}
	myString := string(jsonSendData[:])
	log.Printf("myString = %s", myString)
	payload := strings.NewReader(myString)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", apiKey)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var resultStruct resStruct

	err := json.Unmarshal(body, &resultStruct)
	if err != nil {
		panic(err)
		return "fault to translate"
	}

	if sendingText == resultStruct.Result {
		return "fault to translate"
	}
	return resultStruct.Result
}

func isCyrillic(text string) bool {
	if text[0] > 191 && text[0] <= 255 {
		return true
	}

	if text[0] == 168 || text[0] == 184 {
		return true
	}
	return false
}
