package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"time"
)

type AcceptFriends struct {
	code string
	api  *Vk.Api
	stop chan bool
}

func GetAcceptFriends(api *Vk.Api) AcceptFriends {
	return AcceptFriends{
		code: `var x = API.friends.getRequests({need_viewed:1}).items;
var i = 0;
while (x.length > 0 && i < 25) {
i = i + 1;
var id = x.pop();
API.friends.add({user_id:id});
}
return i;`,
		stop: make(chan bool, 1),
		api: api,
	}
}

func (ac *AcceptFriends) Execute() error {
	err := ac.api.BlindExecute(ac.code)
	if err != nil {
		return err
	}
	return nil
}

func (ac *AcceptFriends) Start() error {
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick:
			err := ac.Execute()
			if err != nil {
				return err
			}
		case <-ac.stop:
			return nil
		}
	}
}

func (ac *AcceptFriends) GetName() string {
	return "AcceptFriends"
}


func (ac *AcceptFriends) Stop() {
	ac.stop <- true
}
