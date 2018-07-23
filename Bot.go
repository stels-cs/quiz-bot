package main

import (
	"log"
	"fmt"
	"sync"
	"strconv"
	"github.com/stels-cs/vk-api-tools"
	"strings"
)

const helpMessageTmp = `Я — бот для игры в квиз.

Разреши мне читать все сообщения в этой беседе чтобы начать играть

Команды: 
@%s %s
@%s %s
@%s %s
`
const startGameCommand = `начать игру`
const stopGameCommand = `закончить игру`
const topCommand = `рейтинг`

type Bot struct {
	logger         *log.Logger
	MyId           int
	Games          map[int]*Game
	db             *QuestionPoll
	top            *Top
	userPoll       *UserPoll
	stopMutex      *sync.Mutex
	stop           chan bool
	captcha        map[int]bool
	deleteGameChan chan int
	TestMode       bool
	queue          *VkApi.RequestQueue
	lastLsId       int
	screenName     string
	groupId        int
	env            string
}

func GetNewBot(queue *VkApi.RequestQueue, logger *log.Logger, poll *QuestionPoll, top *Top, up *UserPoll, name string, id int, enviroment string) Bot {
	b := Bot{
		logger:         logger,
		Games:          map[int]*Game{},
		db:             poll,
		top:            top,
		userPoll:       up,
		stopMutex:      &sync.Mutex{},
		stop:           make(chan bool, 1),
		deleteGameChan: make(chan int, 10),
		captcha:        map[int]bool{},
		queue:          queue,
		screenName:     name,
		groupId:        id,
		env:            enviroment,
	}
	return b
}

func (bot *Bot) GetHelpMessage() string {
	return fmt.Sprintf(helpMessageTmp,
		bot.screenName, startGameCommand,
		bot.screenName, stopGameCommand,
		bot.screenName, topCommand,
	)
}

func (bot *Bot) IsMeMention(text string) bool {

	if strings.Index(text, "["+bot.screenName+"|") == 0 {
		return true
	}

	if strings.Index(text, "[public"+strconv.Itoa(bot.groupId)+"|") == 0 {
		return true
	}

	if strings.Index(text, "[club"+strconv.Itoa(bot.groupId)+"|") == 0 {
		return true
	}

	if strings.Index(text, "[event"+strconv.Itoa(bot.groupId)+"|") == 0 {
		return true
	}

	return false
}

func (bot *Bot) IsStartMessage(text string) bool {
	ptr := []string{startGameCommand, "го", "go", "играть", "начать", "yfxfnm buhe"}
	if strings.Index(text, "[") != 0 {
		return false
	}
	i := strings.Index(text, "]")
	for _, word := range ptr {
		if i != -1 && strings.Index(text, word) >= i {
			return true
		}
	}
	return false
}

func (bot *Bot) IsStopMessage(text string) bool {
	ptr := []string{stopGameCommand, "stop", "стоп", "stop", "pfrjyxbnm buhe"}
	if strings.Index(text, "[") != 0 {
		return false
	}
	i := strings.Index(text, "]")
	for _, word := range ptr {
		if i != -1 && strings.Index(text, word) >= i {
			return true
		}
	}
	return false
}

func (bot *Bot) IsTopMessage(text string) bool {
	ptr := []string{topCommand, "победители", "htqnbyu"}
	if strings.Index(text, "[") != 0 {
		return false
	}
	for _, word := range ptr {
		if strings.Index(text, word) != -1 {
			return true
		}
	}
	return false
}

func (bot *Bot) NewMessage(msg *VkApi.CallbackMessage) {
	isDialog := msg.PeerId > 2e9
	userId := msg.PeerId

	if !isDialog {
		if bot.lastLsId != userId {
			bot.lastLsId = userId
			bot.Say(userId, "Я работаю только в беседах, добавь меня в беседу.")
		}
		return
	}

	if msg.Out == 1 {
		return
	}

	if bot.IsMeMention(trimAndLower(msg.Text)) {
		if bot.IsStartMessage(trimAndLower(msg.Text)) {
			if _, ok := bot.Games[userId]; ok == false {
				game := GetNewGame(userId, bot.queue, bot.db, bot.top, bot.userPoll, bot.logger)
				bot.Games[ userId ] = game
				go func() {
					game.Start()
					bot.deleteGameChan <- game.peerId
				}()
				bot.logger.Printf(fmt.Sprintf("Start game by id%d", userId))
			}
		} else if bot.IsStopMessage(trimAndLower(msg.Text)) {
			if game, ok := bot.Games[userId]; ok && game != nil {
				game.Stop(true)
				bot.logger.Printf(fmt.Sprintf("Stop game by id%d", userId))
			}
		} else if bot.IsTopMessage(trimAndLower(msg.Text)) {
			bot.Say(userId, bot.GetTopString())
		} else {
			bot.Say(userId, bot.GetHelpMessage())
		}
	} else {
		if g, ok := bot.Games[userId]; ok {
			g.Message(msg)
		}
	}
}

func (bot *Bot) Say(peerId int, message string) {
	go func() {
		r := <-bot.queue.Call(VkApi.GetApiMethod("messages.send", VkApi.P{
			"peer_id": strconv.Itoa(peerId),
			"message": message,
		}))
		if r.Err != nil {
			bot.logger.Println(r.Err.Error())
		}
	}()
}

func (bot *Bot) Start() error {
	go bot.queue.Start()
	for {
		select {
		case <-bot.stop:
			return nil
		case peerId := <-bot.deleteGameChan:
			delete(bot.Games, peerId)
		}
	}
}

func (bot *Bot) onEvent(event *VkApi.CallbackEvent) {
	if event.IsMessage() {
		msg, err := event.GetMessage()
		if err != nil {
			bot.logger.Println("Cant get message from event: " + err.Error())
			bot.logger.Println(string(event.Object))
			return
		}
		if bot.env != "production" {
			bot.logger.Println(string(event.Object))
		}
		bot.NewMessage(msg)
	} else {
		bot.logger.Println("Event: " + event.Type)
		bot.logger.Println(string(event.Object))
	}
}

func (bot *Bot) Stop() {
	bot.queue.Stop()
	bot.stop <- true
}

func (bot *Bot) GetTopString() string {

	top := bot.top.GetTop10()
	str := ""

	var uIds []int

	for _, v := range top {
		uIds = append(uIds, v[1])
	}

	users := bot.userPoll.Get(uIds)

	for i := 0; i < 10; i++ {
		if top[i][1] > 0 {
			str += fmt.Sprintf("%d %s – %s\n", top[i][0], transChoose(top[i][0], "балл", "балла", "баллов"), users[top[i][1]].FirstName+" "+users[top[i][1]].LastName)
		}
	}

	if str == "" {
		str = "Никто еще ничего не угадывал (("
	}

	return str
}

func (bot *Bot) GetName() string {
	return "QuizBot"
}
