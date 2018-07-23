package main

import (
	"os"
	"log"
	"os/signal"
	"syscall"
	"time"
	"math/rand"
	"github.com/stels-cs/vk-api-tools"
	"strconv"
)

const ApiVersion = "5.75"

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	rand.Seed(time.Now().UnixNano())

	apiToken := env("VK_TOKEN", "")

	enviroment := env("ENV", "debug")

	if apiToken == "" {
		panic("No token passed, pass token in VK_TOKEN environment")
	}

	qp := QuestionPoll{}
	err := qp.LoadFromFile(env("QUESTION_FILE", "quiz.txt"))
	if err != nil {
		panic(err.Error())
	}

	top := GetTop(env("TOP_FILE", "top.txt"), logger)
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

	logger.Printf("Starting bot with token: "+trimToken(apiToken))

	signalChan := make(chan os.Signal, 1)

	api := VkApi.CreateApi(apiToken, ApiVersion, VkApi.GetHttpTransport(), 30)
	group, err := api.Call("groups.getById", VkApi.P{"fields":"screen_name"})
	if err != nil {
		logger.Println("Called groups.getById, but got error, perhaps bad token, exiting....")
		panic(err)
	}
	groupId := group.QIntDef("0.id", 0)
	groupName := group.QStringDef("0.name", "DELETED")
	groupScreen := group.QStringDef("0.screen_name", "club")
	if groupId == 0 {
		logger.Println("Called groups.getById, but got no group, perhaps api changes exiting....")
		panic("Cant get group info by call groups.getById with token: " + trimToken(apiToken) )
	}

	res, err:= api.Call("groups.setLongPollSettings", VkApi.P{
		"group_id" : strconv.Itoa(groupId),
		"enabled" : "1",
		"api_version": "5.90",
		"message_new": "1",
		"message_reply": "0",
		"photo_new": "0",
		"audio_new": "0",
		"video_new": "0",
		"wall_reply_new": "0",
		"wall_reply_edit": "0",
		"wall_reply_delete": "0",
		"wall_reply_restore": "0",
		"wall_post_new": "0",
		"board_post_new": "0",
		"board_post_edit": "0",
		"board_post_restore": "0",
		"board_post_delete": "0",
		"photo_comment_new": "0",
		"photo_comment_edit": "0",
		"photo_comment_delete": "0",
		"photo_comment_restore": "0",
		"video_comment_new": "0",
		"video_comment_edit": "0",
		"video_comment_delete": "0",
		"video_comment_restore": "0",
		"market_comment_new": "0",
		"market_comment_edit": "0",
		"market_comment_delete": "0",
		"market_comment_restore": "0",
		"poll_vote_new": "0",
		"group_join": "0",
		"group_leave": "0",
		"group_change_settings": "0",
		"group_change_photo": "0",
		"group_officers_edit": "0",
		"message_allow": "1",
		"message_deny": "1",
		"wall_repost": "0",
		"user_block": "0",
		"user_unblock": "0",
		"messages_edit": "0",
		"message_typing_state": "0",
	})

	if err != nil {
		logger.Println("Cant call groups.setLongPollSettings, perhaps token without manage rights")
		panic(err)
	}

	if res.IntDef(0) != 1 {
		logger.Println("Cant call groups.setLongPollSettings (result is not 1), perhaps token without manage rights")
		panic("Cant call groups.setLongPollSettings^ return: "+res.String())
	}

	logger.Println("Long poll settings setted")


	queue := VkApi.GetRequestQueue(api, 20)

	logger.Println("Starting bot at group: "+groupName + " #"+ strconv.Itoa( groupId ) + " " + groupScreen)
	lp := VkApi.GetBotLongPollServer(api, logger)
	userPoll := GetPoll(api, logger)
	bot := GetNewBot(queue, logger, &qp, &top, &userPoll, groupScreen, groupId, enviroment)
	lp.SetListener(bot.onEvent)

	services := GetServicePoll(logger)
	services.Push(&bot)
	services.Push(lp)

	services.RunAll()

	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGUSR1)
	sig := <-signalChan
	logger.Println(sig.String())
	logger.Println("Stoping...")
	top.Save()
	<- services.StopAll()
	logger.Println("Done!")
}
