package main

import (
	"log"
	"strconv"
	"time"
)

type GameMessage struct {
	UserId int
	Text   string
}

type Game struct {
	peerId                  int
	db                      *QuestionPoll
	question                *Question
	questionWaitTime        int
	ignoredQuestion         int
	wasMessageAfterQuestion bool

	message       chan *GameMessage
	stop          chan bool
	timer         *time.Timer
	lastWinUserId int
	winCount      int
	userPoll      *UserPoll
	logger        *log.Logger
	name          string

	onSay          func(msg string)
	onEnd          func(msg string)
	onUserGotPoint func(userId int) int

	onQuestionGot func()
	onQuestion    func()

	generateQuestionTime time.Time
}

func GetNewGame(peerId int, lp *QuestionPoll, up *UserPoll, logger *log.Logger) *Game {
	return &Game{
		peerId:   peerId,
		message:  make(chan *GameMessage, 100),
		stop:     make(chan bool, 10),
		db:       lp,
		userPoll: up,
		logger:   logger,
	}
}

func (game *Game) onMessage(userId int, text string) {
	game.wasMessageAfterQuestion = true
	game.ignoredQuestion = 0
	text = trimAndLower(text)

	godMod := userId == INUserId && text == "да этого никто не знает"

	if text == game.question.Answer || godMod {
		game.onQuestionGot()
		game.timer.Stop()
		if game.lastWinUserId != userId {
			game.winCount = 0
			game.lastWinUserId = userId
		}
		game.winCount++

		game.NewQuestion(game.getCongratulationText(userId, game.onUserGotPoint(userId), text == game.question.Answer) + "\n\n")
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
		game.onSay(game.getAnswerView())
		game.timer.Reset(10 * time.Second)
	}
}
func (game *Game) NewQuestion(prefix string) {
	game.generateQuestionTime = time.Now()
	game.questionWaitTime = 0
	game.wasMessageAfterQuestion = false
	game.question = game.db.GetQuestion()
	game.onSay(prefix + game.question.GetBuzzyText() + "\n" + game.getAnswerView())
	if game.timer == nil {
		game.timer = time.NewTimer(10 * time.Second)
	} else {
		game.timer.Reset(10 * time.Second)
	}
	game.onQuestion()
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

	if openChars == 0 {
		b := transChoose(len(answer), "буква", "буквы", "букв")
		return o + "  " + strconv.Itoa(len(answer)) + " " + b
	} else {
		return o
	}
}

func (game *Game) getUserNme(id int) string {
	u := game.userPoll.Get([]int{id})
	return u[id].FirstName + " " + u[id].LastName + ff(u[id].Sex == 1, " права, у неё уже", " прав, у него уже")
}

func (game *Game) getCongratulationText(userId int, point int, fullMath bool) string {
	str := game.getUserNme(userId) + " " + strconv.Itoa(point) + " " + transChoose(point, "балл", "балла", "баллов")
	if !fullMath {
		str = game.question.Answer + "\n" + str
	}
	return str
}

func (game *Game) Start() {
	game.NewQuestion("Погнали\n\n")
	for {
		select {
		case normalStop := <-game.stop:
			if normalStop {
				if game.timer != nil {
					game.timer.Stop()
				}
				game.onEnd("Игра закончена")
			}
			return
		case msg := <-game.message:
			game.onMessage(msg.UserId, msg.Text)
		case <-game.timer.C:
			game.onTimeout()
		}
	}
}

func (game *Game) Stop(correctStop bool) {
	game.stop <- correctStop
}

func (game *Game) Message(userId int, text string) {
	game.message <- &GameMessage{UserId: userId, Text: text}
}
