package main

import (
	"encoding/json"
	"fmt"
	"github.com/stels-cs/vk-api-tools"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const helpMessageTmp = `Я – бот для игры в квиз.

Разреши мне читать все сообщения в этой беседе, чтобы начать играть. А как это сделать и как вообще играть прочитай вот тут: vk.com/@vikobot-bot

Команды:
@%s %s
@%s %s
@%s %s
@%s %s
`
const startGameCommand = `начать игру`
const stopGameCommand = `закончить игру`
const topCommand = `рейтинг`
const helpCommand = `помощь`

const DevChatId = 2000000004
const INUserId = 19039187

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

	maxPeerId          int
	msgCount           int
	floodCount         int
	gotQuestionCount   int
	totalQuestionCount int
	totalGameCount     int
	startTime          time.Time
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
		startTime:      time.Now(),
		maxPeerId:      2e9,
		msgCount:       0,
	}
	return b
}

func (bot *Bot) GetHelpMessage() string {
	return fmt.Sprintf(helpMessageTmp,
		bot.screenName, startGameCommand,
		bot.screenName, stopGameCommand,
		bot.screenName, topCommand,
		bot.screenName, helpCommand,
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

func (bot *Bot) IsCommand(text string, command string) bool {
	ptr := []string{command}
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

func (bot *Bot) NewMessage(msg *VkApi.CallbackMessage) {
	isDialog := msg.PeerId > 2e9
	peerId := msg.PeerId
	bot.msgCount++
	if peerId > bot.maxPeerId {
		bot.maxPeerId = peerId
	}

	if msg.PeerId == DevChatId && bot.env == "production" {
		return
	}

	if msg.PeerId != DevChatId && bot.env != "production" {
		return
	}

	if !isDialog {
		if bot.lastLsId != peerId {
			bot.lastLsId = peerId
			bot.SayNoKbd(peerId, "Я работаю только в беседах, добавь меня в беседу.")
		}
		return
	}

	if msg.Out == 1 {
		return
	}

	if bot.IsMeMention(trimAndLower(msg.Text)) {
		if bot.IsStartMessage(trimAndLower(msg.Text)) {
			if _, ok := bot.Games[peerId]; ok == false {
				game := GetNewGame(peerId, bot.db, bot.top, bot.userPoll, bot.logger, bot.screenName)
				bot.Games[peerId] = game
				game.onSay = func(msg string) {
					bot.SayStopKbd(peerId, msg)
				}
				game.onEnd = func(msg string) {
					bot.Say(peerId, msg)
				}
				game.onQuestionGot = func() {
					bot.gotQuestionCount++
				}
				game.onQuestion = func() {
					bot.totalQuestionCount++
				}
				go func() {
					game.Start()
					bot.deleteGameChan <- game.peerId
				}()
				bot.logger.Printf(fmt.Sprintf("Start game by id%d", peerId))
				bot.totalGameCount++
			}
		} else if bot.IsStopMessage(trimAndLower(msg.Text)) {
			if game, ok := bot.Games[peerId]; ok && game != nil {
				game.Stop(true)
				bot.logger.Printf(fmt.Sprintf("Stop game by id%d", peerId))
			}
		} else if strings.Index(trimAndLower(msg.Text), "bstat") != -1 {
			bot.Say(peerId, fmt.Sprintf(`Stat:
Total message: %d
Total questions: %d (%d)
Flood control: %d
Dialog count: %d
Games count: %d
Top count: %d
Start at: %s
%s
`,
				bot.msgCount,
				bot.totalQuestionCount,
				bot.gotQuestionCount,
				bot.floodCount,
				bot.maxPeerId-2e9,
				bot.totalGameCount,
				len(bot.top.data),
				bot.startTime.Format("Mon Jan _2 15:04:05"),
				bot.top.GetFastUsers()))
		} else if bot.IsTopMessage(trimAndLower(msg.Text)) {
			bot.Say(peerId, bot.GetTopString())
		} else if bot.IsCommand(msg.Text, "CLEAR") && msg.FromId == INUserId {
			cl := strings.Index(msg.Text, "CLEAR")
			if cl != -1 {
				dig := strings.TrimSpace(string([]rune(msg.Text)[cl+5:]))
				userId, err := strconv.Atoi(dig)
				if err == nil {
					bot.top.Clear(userId)
					bot.Say(peerId, "DONE")
				} else {
					bot.Say(peerId, err.Error())
				}
			}
		} else {
			bot.Say(peerId, bot.GetHelpMessage())
		}
	} else {
		if g, ok := bot.Games[peerId]; ok {
			g.Message(msg.FromId, msg.Text)
		}
	}
}

func (bot *Bot) SayKBD(peerId int, message string, kdb *Keyboard) {
	go func() {

		rawKbd, err := json.Marshal(kdb)
		if err != nil {
			bot.logger.Println(err)
			return
		}

		r := <-bot.queue.Call(VkApi.GetApiMethod("messages.send", VkApi.P{
			"peer_id":  strconv.Itoa(peerId),
			"message":  message,
			"keyboard": string(rawKbd),
		}))
		if r.Err != nil {
			bot.logger.Println(r.Err.Error())
			if strings.Index(r.Err.Error(), "Flood control") != -1 {
				bot.floodCount++
			}
		}
	}()
}

func (bot *Bot) Say(peerId int, message string) {
	go func() {

		keyboard, err := GetDefaultKbd()
		if err != nil {
			bot.logger.Println(err)
			return
		}

		bot.SayKBD(peerId, message, keyboard)
	}()
}

func (bot *Bot) SayStopKbd(peerId int, message string) {
	go func() {

		keyboard, err := GetStopKbd()
		if err != nil {
			bot.logger.Println(err)
			return
		}

		bot.SayKBD(peerId, message, keyboard)
	}()
}

func (bot *Bot) SayNoKbd(peerId int, message string) {
	go func() {
		r := <-bot.queue.Call(VkApi.GetApiMethod("messages.send", VkApi.P{
			"peer_id": strconv.Itoa(peerId),
			"message": message,
		}))
		if r.Err != nil {
			bot.logger.Println(r.Err.Error())
			if strings.Index(r.Err.Error(), "Flood control") != -1 {
				bot.floodCount++
			}
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

	for peerId := range bot.Games {
		bot.Say(peerId, "Игра прервана потому что мы обновлем бота, пожалуйста, напишите боту через несколько секунд.")
	}

	time.Sleep(1 * time.Second)
	bot.queue.Stop()
	time.Sleep(1 * time.Second)
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
			str += fmt.Sprintf(
				"%d %s – @id%d (%s %s)\n",
				top[i][0],
				transChoose(top[i][0], "балл", "балла", "баллов"),
				users[top[i][1]].Id, users[top[i][1]].FirstName, users[top[i][1]].LastName)
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
