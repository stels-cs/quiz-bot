package main

import (
	"github.com/stels-cs/vk-api-tools"
	"log"
	"strconv"
	"strings"
)

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Sex       int    `json:"sex"`
}

type UserPoll struct {
	poll   map[int]User
	api    *VkApi.Api
	logger *log.Logger
}

func GetPoll(api *VkApi.Api, logger *log.Logger) *UserPoll {
	return &UserPoll{map[int]User{}, api, logger}
}

func (up *UserPoll) Get(userIds []int) map[int]User {
	result := map[int]User{}
	var toRequest []int
	for _, v := range userIds {
		if u, ok := up.poll[v]; ok == false {
			toRequest = append(toRequest, v)
			result[v] = User{Id: v, FirstName: "id" + strconv.Itoa(v), LastName: ""}
		} else {
			result[v] = u
		}
	}

	users := make([]User, 0)
	if len(toRequest) > 0 {
		err := up.api.Exec("users.get", VkApi.P{
			"user_ids": idsToString(toRequest),
			"fields":   "sex",
		}, &users)
		if err == nil {
			for _, v := range users {
				result[v.Id] = v
				up.poll[v.Id] = v
			}
		} else {
			up.logger.Println(err)
		}
	}

	return result
}

func idsToString(ids []int) string {
	s := make([]string, 0, len(ids))
	for _, id := range ids {
		s = append(s, strconv.Itoa(id))
	}
	return strings.Join(s, ",")
}
