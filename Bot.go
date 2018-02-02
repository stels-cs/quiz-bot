package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"log"
	"fmt"
	"sync"
	"strconv"
)

const helpMessage = `Я — бот для игры в квиз.

Отправь мне заявку в друзья, и я автоматически её приму. Затем добавляй меня в чатик с друзьями и играй.

Список команд:
бот начать – начать игру.
бот стоп – завершить игру (игра завершается автоматически, если в течение 3 вопросов в чате не было ни одного сообщения).
бот рейтинг – показать топ-10 пользователей по количеству правильных ответов.
бот помощь – это сообщение.

Если бот не отвечает, то введите капчу тут %s
`

const TestPeerId = 2000000001

type Bot struct {
	Vk.LongPollDefaultListener
	logger            *log.Logger
	token             Vk.AccessToken
	MyId              int
	Games             map[int]*Game
	db                *QuestionPoll
	top               *Top
	userPoll          *UserPoll
	stopMutex         *sync.Mutex
	restoreUrl        string
	stop              chan bool
	api               *Vk.Api
	lpServer          *Vk.LongPollServer
	captcha           map[int]bool
	deleteGameChan    chan int
	captchaReportChan chan int
	TestMode          bool
	queue             *Vk.RequestQueue
	lastLsId	int
}

func GetNewBot(lpServer *Vk.LongPollServer, queue *Vk.RequestQueue, myId int, logger *log.Logger, poll *QuestionPoll, top *Top, up *UserPoll, restoreUrl string, test bool) Bot {
	b := Bot{
		logger:            logger,
		lpServer:          lpServer,
		MyId:              myId,
		Games:             map[int]*Game{},
		db:                poll,
		top:               top,
		userPoll:          up,
		stopMutex:         &sync.Mutex{},
		restoreUrl:        restoreUrl,
		stop:              make(chan bool, 1),
		deleteGameChan:    make(chan int, 10),
		captchaReportChan: make(chan int, 10),
		captcha:           map[int]bool{},
		TestMode:          test,
		queue:             queue,
	}
	lpServer.SetListener(&b)
	return b
}

func (bot *Bot) NewMessage(msg Vk.MessageEvent) {
	if msg.From == 0 {
		if bot.lastLsId != msg.PeerId {
			bot.lastLsId = msg.PeerId
			bot.Say(msg.PeerId, "Я работаю только в беседах, добавь меня в беседу.")
		}
		return
	}
	if msg.From == bot.MyId || msg.From == 0 {
		return
	}

	if bot.TestMode && msg.PeerId != TestPeerId {
		return
	} else if !bot.TestMode && msg.PeerId == TestPeerId {
		return
	}

	if msg.ChatInviteUser != nil && msg.ChatInviteUser.User == bot.MyId {
		bot.logger.Printf(fmt.Sprintf("Invite to chat by id%d", msg.From))
		bot.Say(msg.PeerId, fmt.Sprintf(helpMessage, bot.restoreUrl))
	} else if msg.ChatKilUser != nil && msg.ChatKilUser.User == bot.MyId {
		bot.logger.Printf(fmt.Sprintf("Kick from chat by id%d", msg.From))
		if game, ok := bot.Games[msg.PeerId]; ok == true {
			game.Stop(false)
		}
	} else if inArray([]string{"бот помощ", "бот помощь", "бот помошь", "бот хелп", "бот help", "/help", "!help", "bot help", "bot /help", "bot !help"}, trimAndLower(msg.Text)) {
		bot.Say(msg.PeerId, fmt.Sprintf(helpMessage, bot.restoreUrl))
	} else if inArray([]string{"бот начать", "го квиз", "бот старт"}, trimAndLower(msg.Text)) {
		if _, ok := bot.Games[msg.PeerId]; ok == false {
			game := GetNewGame(msg.PeerId, bot.queue, bot.db, bot.top, bot.userPoll, bot.logger)
			bot.Games[ msg.PeerId ] = game
			go func() {
				correctStop := game.Start(bot.restoreUrl)
				if !correctStop {
					bot.captchaReportChan <- game.peerId
				}
				bot.deleteGameChan <- game.peerId
			}()
			bot.logger.Printf(fmt.Sprintf("Start game by id%d", msg.From))
		}
	} else if inArray([]string{"бот стоп", "bot stop",}, trimAndLower(msg.Text)) {
		if game, ok := bot.Games[msg.PeerId]; ok && game != nil {
			game.Stop(true)
			bot.logger.Printf(fmt.Sprintf("Stop game by id%d", msg.From))
		}
	} else if inArray([]string{"бот рейтинг", "бот топ"}, trimAndLower(msg.Text)) {
		bot.Say(msg.PeerId, bot.GetTopString())
	} else if inArray([]string{"captcha.force"}, trimAndLower(msg.Text)) && bot.TestMode {
		bot.logger.Println("captcha.force")
		bot.CaptchaForce(msg.PeerId)
	} else {
		if g, ok := bot.Games[msg.PeerId]; ok {
			g.Message(&msg)
		} else {
			//bot.logger.Printf("%+v\n", msg)
		}
	}
}

func (bot *Bot) Say(peerId int, message string) {
	go func() {
		r := <-bot.queue.Call(Vk.GetApiMethod("messages.send", Vk.Params{
			"peer_id": strconv.Itoa(peerId),
			"message": message,
		}))
		if r.Err != nil {
			bot.logger.Println(Vk.PrintError(r.Err))
			if apiErr, ok := r.Err.(*Vk.ApiError); ok && apiErr.Code == Vk.ApiErrorCaptcha {
				bot.captchaReportChan <- peerId
			}
		}
	}()
}

func (bot *Bot) CaptchaForce(peerId int) {
	go func() {
		r := <-bot.queue.Call(Vk.GetApiMethod("captcha.force", Vk.Params{}))
		if r.Err != nil {
			bot.logger.Println(Vk.PrintError(r.Err))
			if apiErr, ok := r.Err.(*Vk.ApiError); ok && apiErr.Code == Vk.ApiErrorCaptcha {
				bot.captchaReportChan <- peerId
			}
		}
	}()
}

func (bot *Bot) EditMessage(msg Vk.MessageEvent) {
	bot.NewMessage(msg)
}

func (bot *Bot) Start() error {
	go bot.queue.Start()
	go bot.lpServer.Start()
	for {
		select {
		case <-bot.stop:
			return nil
		case peerId := <-bot.deleteGameChan:
			delete(bot.Games, peerId)
		case peerId := <-bot.captchaReportChan:
			if peerId == -1 {
				bot.captchaRecover()
			} else {
				bot.captcha[peerId] = true
			}
		}
	}
}

func (bot *Bot) Stop() {
	bot.queue.Stop()
	bot.lpServer.Stop()
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

func (bot *Bot) OnCaptchaRecover() {
	bot.captchaReportChan <- -1
}

func (bot *Bot) captchaRecover() {
	for peerId := range bot.captcha {
		bot.Say(peerId, "Бот не отвечал из-за капчи, но теперь все хорошо, скажи \"го квиз\" и поиграем")
	}
	bot.captcha = map[int]bool{}
}

func (bot *Bot) GetName() string {
	return "QuizBot"
}
