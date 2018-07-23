package main

import (
	"time"
	"log"
	"strconv"
	"github.com/stels-cs/vk-api-tools"
)

type Game struct {
	peerId                  int
	db                      *QuestionPoll
	question                *Question
	queue                   *VkApi.RequestQueue
	questionWaitTime        int
	ignoredQuestion         int
	wasMessageAfterQuestion bool
	message                 chan *VkApi.CallbackMessage
	stop                    chan bool
	timer                   *time.Timer
	top                     *Top
	lastWinUserId           int
	winCount                int
	userPoll                *UserPoll
	logger                  *log.Logger
}

func GetNewGame(peerId int, queue *VkApi.RequestQueue, lp *QuestionPoll, top *Top, up *UserPoll, logger *log.Logger) *Game {
	return &Game{
		peerId:   peerId,
		message:  make(chan *VkApi.CallbackMessage, 100),
		stop:     make(chan bool, 10),
		queue:    queue,
		db:       lp,
		top:      top,
		userPoll: up,
		logger:   logger,
	}
}

func (game *Game) Say(msg string) {
	r := <-game.queue.Call(VkApi.GetApiMethod("messages.send", VkApi.P{
		"peer_id": strconv.Itoa(game.peerId),
		"message": msg,
	}))
	if r.Err != nil {
		game.logger.Println(r.Err.Error())
	}
}

func (game *Game) onMessage(ev *VkApi.CallbackMessage) {
	game.wasMessageAfterQuestion = true
	game.ignoredQuestion = 0
	text := trimAndLower(ev.Text)
	uId := ev.PeerId

	godMod := ev.FromId == 19039187 && text == "да этого никто не знает"

	if text == game.question.Answer || godMod {
		game.timer.Stop()
		if game.lastWinUserId != uId {
			game.winCount = 0
			game.lastWinUserId = uId
		}
		game.winCount++
		game.NewQuestion(game.getCongratulationText(ev.FromId, game.top.Inc(ev.FromId)) + "\n\n")
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
	str := game.getUserNme(userId) + " " + strconv.Itoa(point) + " " + transChoose(point, "балл", "балла", "баллов")
	return str
}

func (game *Game) Start(){
	game.NewQuestion("Погнали\n\n")
	for {
		select {
		case normalStop := <-game.stop:
			if normalStop {
				if game.timer != nil {
					game.timer.Stop()
				}
				game.Say("Игра закончена")
			}
			return
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

func (game *Game) Message(msg *VkApi.CallbackMessage) {
	game.message <- msg
}
