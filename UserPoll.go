package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"log"
	"strconv"
)

type UserPoll struct {
	poll map[int]Vk.User
	api *Vk.Api
	logger *log.Logger
}

func GetPoll(api *Vk.Api, logger *log.Logger) UserPoll {
	return UserPoll{ map[int]Vk.User{}, api, logger }
}

func (up *UserPoll) Get( userIds []int) map[int]Vk.User {
	result := map[int]Vk.User{}
	var toRequest []int
	for _,v:=range userIds {
		if u, ok := up.poll[v]; ok == false {
			toRequest = append(toRequest, v)
			result[v] = Vk.User{Id:v,FirstName:"id" + strconv.Itoa(v),LastName:""}
		} else {
			result[v] = u
		}
	}

	if len(toRequest) > 0 {
		users, err := up.api.Users.GetByIds( toRequest )
		if err == nil {
			for _,v:=range users {
				result[v.Id] = v
				up.poll[v.Id] = v
			}
		} else {
			up.logger.Println( Vk.PrintError(err) )
		}
	}

	return result
}
