package main

import (
	"log"
	"fmt"
	"time"
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
	errorCount := 0
	lastEventTime := time.Now()
	go func() {
		for {
			err := service.Start()
			if sp.allStop {
				sp.logger.Println(fmt.Sprintf("[%s] stopped", service.GetName()))
				sp.onStopService()
				return
			} else if err != nil {
				sp.logger.Println(fmt.Sprintf("[%s] %s", service.GetName(), err.Error()))

				errorCount++
				if errorCount > 1000 {
					panic(service.GetName() + " generate too more errors")
				}
				if time.Now().Sub(lastEventTime) > 10 * time.Second {
					errorCount = 0
				}
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
