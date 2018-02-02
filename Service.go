package main

import (
	"log"
	"github.com/stels-cs/quiz-bot/Vk"
	"fmt"
)

type Service interface {
	Start() error
	GetName() string
	Stop()
}

type ServicePoll struct {
	logger   *log.Logger
	poll     []Service
	stop     chan bool
	allStop  bool
	stopWait int
}

func GetServicePoll(looger *log.Logger) ServicePoll {
	return ServicePoll{
		logger: looger,
		poll:   []Service{},
	}
}

func (sp *ServicePoll) Push(service Service) {
	sp.poll = append(sp.poll, service)
}

func (sp *ServicePoll) RunAll() {
	sp.allStop = false
	for _, v := range sp.poll {
		sp.run(v)
	}
}

func (sp *ServicePoll) StopAll() chan bool {
	sp.allStop = true
	sp.stopWait = 0
	for _, v := range sp.poll {
		sp.stopWait++
		sp.logger.Println(fmt.Sprintf("[%s] stopping...", v.GetName()))
		v.Stop()
	}
	sp.stop = make(chan bool, 1)
	return sp.stop
}

func (sp *ServicePoll) run(service Service) {
	sp.logger.Println(fmt.Sprintf("[%s] is started", service.GetName()))
	go func() {
		for {
			err := service.Start()
			if sp.allStop {
				sp.logger.Println(fmt.Sprintf("[%s] stopped", service.GetName()))
				sp.onStopService()
				return
			} else if err != nil {
				sp.logger.Println(fmt.Sprintf("[%s] %s", service.GetName(), Vk.PrintError(err)))
			} else {
				sp.logger.Println(fmt.Sprintf("[%s] Restarted", service.GetName()))
			}
		}
	}()
}

func (sp *ServicePoll) onStopService() {
	sp.stopWait--
	if sp.stopWait <= 0 {
		sp.stop <- true
	}
}
