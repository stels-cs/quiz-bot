package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"time"
	"log"
	"strconv"
	"fmt"
)

type Game struct {
	peerId                  int
	db                      *QuestionPoll
	question                *Question
	queue                   *Vk.RequestQueue
	questionWaitTime        int
	ignoredQuestion         int
	wasMessageAfterQuestion bool
	message                 chan *Vk.MessageEvent
	stop                    chan bool
	captcha                 chan bool
	timer                   *time.Timer
	top                     *Top
	lastWinUserId           int
	winCount                int
	userPoll                *UserPoll
	logger                  *log.Logger
}

func GetNewGame(peerId int, queue *Vk.RequestQueue, lp *QuestionPoll, top *Top, up *UserPoll, looger *log.Logger) *Game {
	return &Game{
		peerId:   peerId,
		message:  make(chan *Vk.MessageEvent, 100),
		stop:     make(chan bool, 10),
		captcha:     make(chan bool, 10),
		queue:    queue,
		db:       lp,
		top:      top,
		userPoll: up,
		logger:   looger,
	}
}

func (game *Game) Say(msg string) {
	r := <-game.queue.Call(Vk.GetApiMethod("messages.send", Vk.Params{
		"peer_id": strconv.Itoa(game.peerId),
		"message": msg,
	}))
	if r.Err != nil {
		game.logger.Println(Vk.PrintError(r.Err))
		if apiErr, ok := r.Err.(*Vk.ApiError); ok && apiErr.Code == Vk.ApiErrorCaptcha {
			game.captcha <- false
		}
	}
}

func (game *Game) onMessage(event *Vk.MessageEvent) {
	game.wasMessageAfterQuestion = true
	game.ignoredQuestion = 0
	text := trimAndLower(event.Text)
	if text == game.question.Answer {
		game.timer.Stop()
		if game.lastWinUserId != event.From {
			game.winCount = 0
			game.lastWinUserId = event.From
		}
		game.winCount++
		game.NewQuestion(game.getCongratulationText(event.From, game.top.Inc(event.From)) + "\n\n")
	}
}
func (game *Game) onTimeout() {
	if game.questionWaitTime == 0 {
		game.questionWaitTime = 1
	} else if game.questionWaitTime == 1 {
		game.questionWaitTime = 3
	} else {
		game.questionWaitTime++
	}
	if game.questionWaitTime > 3 || game.questionWaitTime > len(game.question.Answer)-1 {
		game.onUnAnswerQuestion()
	} else {
		game.Say(game.getAnswerView())
		game.timer.Reset(10 * time.Second)
	}
}
func (game *Game) NewQuestion(prefix string) {
	game.questionWaitTime = 0
	game.wasMessageAfterQuestion = false
	game.question = game.db.GetQuestion()
	game.Say(prefix + game.question.Text + "\n" + game.getAnswerView())
	if game.timer == nil {
		game.timer = time.NewTimer(10 * time.Second)
	} else {
		game.timer.Reset(10 * time.Second)
	}
}
func (game *Game) onUnAnswerQuestion() {
	if game.wasMessageAfterQuestion == false {
		game.ignoredQuestion++
		if game.ignoredQuestion > 3 {
			game.stop <- true
			return
		}
	}
	game.NewQuestion(game.question.Answer + "\n\n")
}

func (game *Game) getAnswerView() string {
	openChars := game.questionWaitTime
	answer := []rune(game.question.Answer)
	o := ""
	for k := range answer {
		if k == 0 && openChars >= 1 {
			o += string(answer[k])
		} else if k == 1 && openChars >= 2 {
			o += string(answer[k])
		} else if k == 0 {
			o += "*"
		} else if openChars == 3 && k == len(answer)-1 {
			o += " " + string(answer[k])
		} else {
			o += " *"
		}
	}
	return o
}

func (game *Game) getUserNme(id int) string {
	u := game.userPoll.Get([]int{id})
	return u[id].FirstName + " " + u[id].LastName + ff(u[id].Sex == 1, " права, у неё уже", " прав, у него уже")
}

func (game *Game) getCongratulationText(userId int, point int) string {
	str := game.getUserNme(userId) + " " + strconv.Itoa(point) + " " + transChoose(point, "бал", "балла", "баллов")
	return str
}

func (game *Game) Start(url string) bool {
	game.NewQuestion(fmt.Sprintf("Погнали, если бот сломался, введите капчу тут %s\n\n", url))
	for {
		select {
		case <- game.captcha:
			return false
		case normalStop := <-game.stop:
			if normalStop {
				if game.timer != nil {
					game.timer.Stop()
				}
				game.Say("Игра закончена")
			}
			return true
		case msg := <-game.message:
			game.onMessage(msg)
		case <-game.timer.C:
			game.onTimeout()
		}
	}
}

func (game *Game) Stop(correctStop bool) {
	game.stop <- correctStop
}

func (game *Game) Message(msg *Vk.MessageEvent) {
	game.message <- msg
}
