package main

import (
	"flag"
	"fmt"
	"github.com/stels-cs/vk-api-tools"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var defaultLogger *log.Logger

func init() {
	rand.Seed(time.Now().UnixNano())
	defaultLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
}

func renderTop(conTop []*UserItem, vkUsers map[int]User) string {
	str := "\n\nТоп этой беседы:\n"

	for _, userItem := range conTop {
		str += fmt.Sprintf(
			"#%d %s %s - %d %s",
			userItem.GetPlace(),
			vkUsers[userItem.Id].FirstName,
			vkUsers[userItem.Id].LastName,
			userItem.GetScore(),
			transChoose(userItem.GetScore(), "балл", "балла", "баллов"),
		)
	}

	return str
}

func main() {

	apiToken := env("VK_TOKEN", "")

	if apiToken == "" {
		panic("No token passed, pass token in VK_TOKEN environment")
	}

	environment := env("ENV", "debug")

	questionPoll := QuestionPoll{}
	err := questionPoll.LoadFromFile(env("QUESTION_FILE", "quiz.txt"))
	if err != nil {
		panic(err.Error())
	}

	dailyTop, err := GetUserState(true, "daily_top.bolt")
	if err != nil {
		panic(err)
	}

	dailyTopCraetedTime := time.Now()

	globalTop, err := GetUserState(true, "global_top.bolt")

	if err != nil {
		panic(err)
	}

	var ip = flag.Int("copyTop", 0, "copy old top, to bold")
	flag.Parse()

	if ip != nil && *ip == 1 {
		top := GetTop(env("TOP_FILE", "top.txt"), defaultLogger)
		err = top.Load()
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				panic(err.Error())
			}
			print("Top file no exists creating... ")
			err := top.Save()
			if err != nil {
				panic(err.Error())
			} else {
				print("ok\n")
			}
		}
		defaultLogger.Printf("Start coping %d items\n", len(top.data))
		dailyTop.DropTop()
		globalTop.DropTop()
		for userId, score := range top.data {
			dailyTop.AddScoreValue(userId, score)
			globalTop.AddScoreValue(userId, score)
		}
		defaultLogger.Printf("Done\n")
		globalTop.SaveForce()
		dailyTop.SaveForce()
		return
	}

	saveTimer := time.NewTicker(20 * time.Minute)
	go func() {
		for {
			<-saveTimer.C
			globalTop.Save()
			globalTop.Clear()

			dailyTop.Save()
			dailyTop.Clear()
		}
	}()

	stat, err := GetStatistic("statistic.bolt")

	if err != nil {
		panic(err)
	}

	defaultLogger.Printf("Starting bot with token: " + trimToken(apiToken))

	core := CreateBotCore(apiToken)
	core.Logger = defaultLogger

	if err := core.bootstrap(); err != nil {
		panic(err)
	}

	defaultLogger.Println("Done! Bot started at group: " + core.GetGroupView())

	conversationTop, err := GetConversationTop("conversation.bolt")
	if err != nil {
		panic(err)
	}

	Games := map[int]*Game{}

	mutex := sync.Mutex{}

	core.BeforeMessage(func(msg *VkApi.CallbackMessage) bool {
		if environment == "debug" {
			if msg.PeerId != 2000000001 {
				return false
			}
		}
		return true
	})

	core.BeforeMessage(func(msg *VkApi.CallbackMessage) bool {
		if dailyTopCraetedTime.Day() != time.Now().Day() {
			dailyTop.DropTop()
			dailyTop.SaveForce()

			globalTop.Save()
			globalTop.Clear()

			stat.PutUsersInTop(globalTop.GetUserSafety(globalTop.tail).GetPlace())

			dailyTopCraetedTime = time.Now()
		}
		return true
	})

	core.BeforeMessage(func(msg *VkApi.CallbackMessage) bool {
		stat.ReceiveMessage()
		if msg.FromChat() {
			stat.PutDialogCount(msg.PeerId - 2e9)
		}
		return true
	})

	core.OnSendMessage(func() {
		stat.SendMessage()
	})

	//core.OnMessage().WithWords([]string{"show_peer"}).Do(func(msg *VkApi.CallbackMessage) *BotResponse {
	//	return SimpleMessageResponse(strconv.Itoa(msg.PeerId))
	//})

	core.OnMessage().
		FromChat().
		NoGroup().
		WithMention().
		WithWords([]string{startGameCommand, "го", "go", "играть", "начать", "yfxfnm buhe"}).
		Do(func(msg *VkApi.CallbackMessage) *BotResponse {
			mutex.Lock()
			defer mutex.Unlock()
			if _, ok := Games[msg.PeerId]; ok == false {
				game := GetNewGame(msg.PeerId, &questionPoll, core.userDB, defaultLogger)
				Games[msg.PeerId] = game
				game.onUserGotPoint = func(userId int) int {
					dailyTop.AddScore(userId)
					globalTop.AddScore(userId)
					err := conversationTop.AddScore(msg.PeerId, userId)
					if err != nil {
						defaultLogger.Println(err)
					}
					return dailyTop.GetUserSafety(userId).GetScore()
				}
				game.onSay = func(text string) {
					core.SendMessage(TextMessage(text).SetKeyboard(GetStopKeyboad()), msg.PeerId)
				}
				game.onEnd = func(text string) {

					conTop, err := conversationTop.GetTop(msg.PeerId)
					uIds := make([]int, 0, len(conTop))
					if err != nil {
						defaultLogger.Println(err)
					} else if len(conTop) > 0 {
						for _, u := range conTop {
							uIds = append(uIds, u.Id)
						}
						vkUsers := core.userDB.Get(uIds)

						text += renderTop(conTop, vkUsers)
					}

					core.SendMessage(TextMessage(text).SetKeyboard(GetDefaultkeyboad()), msg.PeerId)
					err = conversationTop.Remove(msg.PeerId)
					if err != nil {
						defaultLogger.Println(err)
					}
				}
				game.onQuestionGot = func() {
					stat.DoneQuestions()
				}
				game.onQuestion = func() {
					stat.StartQuestion()
					if environment == "debug" && game.question != nil {
						defaultLogger.Println(game.question.Answer)
					}
				}
				go func() {
					game.Start()
					mutex.Lock()
					defer mutex.Unlock()
					delete(Games, msg.PeerId)
				}()

				stat.StartGame()
			}
			return NoReactionResponse()
		})

	core.OnMessage().
		NoGroup().
		FromChat().
		WithMention().
		WithWords([]string{stopGameCommand, "stop", "стоп", "stop", "pfrjyxbnm buhe"}).
		Do(func(msg *VkApi.CallbackMessage) *BotResponse {
			if game, ok := Games[msg.PeerId]; ok && game != nil {
				game.Stop(true)
				return NoReactionResponse()
			} else {
				return nil
			}
		})

	core.OnMessage().
		FromChat().
		WithMention().
		WithWords([]string{topCommand, "победители", "htqnbyu"}).
		Do(func(msg *VkApi.CallbackMessage) *BotResponse {
			var user *UserItem
			var top []*UserItem

			str := ""

			if strings.Index(msg.Text, "all") != -1 {
				user, top = globalTop.GetTop(msg.FromId, 20)
				str += "Глобальный рейтинг, за все время:\n"
			} else {
				user, top = dailyTop.GetTop(msg.FromId, 20)
			}

			var uIds []int
			for _, user := range top {
				uIds = append(uIds, user.Id)
			}
			uIds = append(uIds, user.Id)

			conTop, err := conversationTop.GetTop(msg.PeerId)
			if err != nil {
				defaultLogger.Println(err)
			} else if len(conTop) > 0 {
				for _, u := range conTop {
					uIds = append(uIds, u.Id)
				}
			}

			vkUsers := core.userDB.Get(uIds)
			currentUserInTop := false
			for _, userItem := range top {
				str += fmt.Sprintf(
					"#%d @id%d (%s %s) - %d %s",
					userItem.GetPlace(),
					vkUsers[userItem.Id].Id,
					vkUsers[userItem.Id].FirstName,
					vkUsers[userItem.Id].LastName,
					userItem.GetScore(),
					transChoose(userItem.GetScore(), "балл", "балла", "баллов"),
				)
				if userItem.Id == msg.FromId {
					str += " *\n"
					currentUserInTop = true
				} else {
					str += "\n"
				}
			}

			if !currentUserInTop {
				str += "\n"
				str += fmt.Sprintf(
					"#%d %s %s - %d %s",
					user.GetPlace(),
					vkUsers[user.Id].FirstName,
					vkUsers[user.Id].LastName,
					user.GetScore(),
					transChoose(user.GetScore(), "балл", "балла", "баллов"),
				)
			}

			if str == "" {
				str = "Глобальный рейтинг за текуший день пока пуст."
			} else {
				if strings.Index(msg.Text, "all") == -1 {
					str += "\nРейтинг обнуляется каждый день в 00:00 по Москве"
				}
			}

			if len(conTop) > 0 {
				str += renderTop(conTop, vkUsers)
			}

			return SimpleMessageResponse(str).SetKeyboard(GetDefaultkeyboad())
		})

	core.OnMessage().WithMention().WithWords([]string{"bstat"}).Do(func(msg *VkApi.CallbackMessage) *BotResponse {
		text := "Статистика за последние 7 дней:\n"

		text += "\nВходящий сообщений: " + idsToString(stat.GetTopLine(ReceiveMessages))
		text += "\nИсходящий сообщений: " + idsToString(stat.GetTopLine(SendMessage))
		text += "\nИгр: " + idsToString(stat.GetTopLine(StartGames))
		text += "\nВопросов: " + idsToString(stat.GetTopLine(StartQuestions))
		text += "\nПравильных ответов: " + idsToString(stat.GetTopLine(DoneQuestions))
		text += "\nДиалогов: " + idsToString(stat.GetTopLine(DialogsCount))
		text += "\nПользователей: " + idsToString(stat.GetTopLine(UserInTopCount))
		mutex.Lock()
		text += "\nОнлайн: " + strconv.Itoa(len(Games))
		mutex.Unlock()

		return SimpleMessageResponse(text)
	})

	core.OnMessage().WithMention().Do(func(msg *VkApi.CallbackMessage) *BotResponse {
		return SimpleMessageResponse(fmt.Sprintf(helpMessageTmp,
			core.groupScreenName, startGameCommand,
			core.groupScreenName, stopGameCommand,
			core.groupScreenName, topCommand,
			core.groupScreenName, helpCommand,
		)).SetKeyboard(GetDefaultkeyboad())
	})

	core.OnMessage().FromChat().Do(func(msg *VkApi.CallbackMessage) *BotResponse {
		if game, ok := Games[msg.PeerId]; ok {
			game.Message(msg.FromId, msg.Text)
		}
		return nil
	})

	core.OnMessage().FromUser().NoChat().Do(func(msg *VkApi.CallbackMessage) *BotResponse {
		return SimpleMessageResponse(directMessage)
	})

	oldTsValue, err := stat.GetTsValue()

	if err == nil {
		core.SetOldTsValue(oldTsValue)
	}

	services := GetServicePoll(defaultLogger)
	services.Push(core)
	services.RunAll()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGUSR1)
	sig := <-signalChan
	defaultLogger.Println(sig.String())
	defaultLogger.Println("Stopping...")

	saveTimer.Stop()
	dailyTop.SaveForce()
	globalTop.SaveForce()

	stat.PutUsersInTop(globalTop.GetUserSafety(globalTop.tail).GetPlace())

	mutex.Lock()

	msg := TextMessage("Мы останавливаем бота чтобы обновить его, он вернется через 1 минуту.").SetKeyboard(GetDefaultkeyboad())
	for peerId := range Games {
		core.SendMessage(msg, peerId)
	}

	mutex.Unlock()

	<-services.StopAll()

	stat.SetTsValue(core.GetTsValue())

	defaultLogger.Println("Done!")
}
