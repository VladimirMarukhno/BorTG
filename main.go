package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type bnResp  struct {
	Price float64 `json:"price,string"`
	Code int64 `json:"code"`
}

type gb = map[string]float64
var db = map[int64] gb {}

func main() {
	currency := "руб."
	bot, err := tgbotapi.NewBotAPI("botTokenHere")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Авторизация")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err:= bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		command := strings.Split(update.Message.Text, " ")
		switch command[0] {

		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда!"))
			}

			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _ , ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = gb{}
			}
			db[update.Message.Chat.ID][command[1]] += amount

			balanceText := fmt.Sprintf("%.2f", db[update.Message.Chat.ID][command[1]] )
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))

		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда!"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _ , ok := db[update.Message.Chat.ID][command[1]]; !ok {
				continue
			}
			if db[update.Message.Chat.ID][command[1]]< amount {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Баланс не может быть меньше нуля!"))
				continue
			}
			db[update.Message.Chat.ID][command[1]] -= amount
			balanceText := fmt.Sprintf("%.2f", db[update.Message.Chat.ID][command[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))

		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда!"))
			}

			delete(db[update.Message.Chat.ID], command[1])

		case "SHOW":
			msg := ""
			var sum float64 = 0
			for key, value := range db[update.Message.Chat.ID] {
				price, _ := getPrice(key)
				sum += value*price
				msg += fmt.Sprintf("%s: %.2f[%.2f %s]\n ", key, value,value*price,currency)
		}
			msg += fmt.Sprintf("total: %.2f %s\n ", sum,currency)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		case "/start" , "/help":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Доступные команды.\nADD - Добавление валюты в кошелёк." +
				"\nDEL - Удаление валюты из кошелька.\nSUB - Уменьшение баланса кошелька.\nSHOW - Проверить баланс кошелька."))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не найдена!"))
		}
	}
}
func getPrice(symbol string) (price float64, err error) {
	resp, err :=http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sRUB", symbol))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp bnResp

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		err = errors.New("неверный символ")
	}

	price = jsonResp.Price
	return
}
