package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"os"
	"log"
	"os/signal"
	"syscall"
	"time"
	"math/rand"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	rand.Seed(time.Now().UnixNano())

	apiToken, err := getToken(env("VK_LOGIN", ""), env("VK_PASSWORD", ""), env("TOKEN_FILE", "token.txt"))
	if err != nil {
		panic(Vk.PrintError(err))
	}

	qp := QuestionPoll{}
	err = qp.LoadFromFile(env("QUESTION_FILE", "quiz.txt"))
	if err != nil {
		panic(err.Error())
	}

	top := GetTop(env("TOP_FILE", "top.txt"), logger)
	err = top.Load()
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			panic(err.Error())
		}
		//Top file not exists try create
		print("Top file no exists creating... ")
		err := top.Save()
		if err != nil {
			panic(err.Error())
		} else {
			print("ok\n")
		}
	}

	logger.Printf("Start: userId: %d token: %s\n", apiToken.UserId, apiToken.Token)

	signalChan := make(chan os.Signal, 1)

	api := Vk.GetApi(apiToken, Vk.GetHttpTransport(), logger)
	queue := Vk.GetRequestQueue(api)
	userPoll := GetPoll(api, logger)
	testMode := true
	if env("ENV", "DEV") == "PRODUCTION" {
		testMode = false
	}
	lpServer := Vk.GetLongPollServer(apiToken, logger)
	bot := GetNewBot(lpServer, queue, apiToken.UserId, logger, &qp, &top, &userPoll, env("RESTORE_URL", "http://localhost:8080/"), testMode)
	deadHand := GetDeadHand(api, &bot)
	friends := GetAcceptFriends(api)

	services := GetServicePoll(logger)
	services.Push(deadHand)
	services.Push(&bot)
	services.Push(&friends)

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
