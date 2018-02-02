package main

import (
	"github.com/stels-cs/quiz-bot/Vk"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"net/http"
	"time"
	"bytes"
	"encoding/json"
)

func ff(cond bool, t string, f string) string {
	if cond {
		return t
	} else {
		return f
	}
}

func env(name string, def string) string {
	v := os.Getenv(name)
	return ff( v!="", v, def )
}

func fillAccessToken( token string ) (Vk.AccessToken, error) {
	t := Vk.AccessToken{ Token:token }
	api := Vk.GetApi(t, Vk.GetHttpTransport(), nil)
	u, err := api.Users.GetMe()
	if err != nil {
		return t, err
	} else {
		t.UserId = u.Id
		return t, nil
	}
}

func getToken(login string, password string, cacheFile string) (Vk.AccessToken, error)  {
	print("Load access token from ", cacheFile, "\n")
	dat, err := ioutil.ReadFile( cacheFile )
	if err != nil {
		print("Error ", err.Error(), "\n")
	} else {
		print("Got token ", string(dat[:10]), "....", "\n")
		print("Checking.... ")
		t, e:= fillAccessToken(string(dat))
		if e != nil {
			print("Fail checking ", e.Error(), "\n")
		} else {
			print("Ok", "\n")
			return t,nil
		}
	}

	print("Getting new access token for login ", login, "\n")

	token := Vk.AccessToken{}
	if login != "" && password != "" {
		token, err = Vk.PasswordAuth(login, password)
		if err != nil {
			return token, err
		} else {
			err := ioutil.WriteFile(cacheFile, []byte(token.Token), 0500)
			if err != nil {
				print("Error when cache access_token ",err.Error())
			}
			return token, nil
		}
	} else {
		return token, errors.New("No token in cache file, no env VK_LOGIN or VK_PASSWORD passed for create new token\n")
	}
}

func transChoose(x int, one string, two string, five string) string {
	if x == 0 {
		return five
	}
	if x > 20 || x < 10 {
		x = x % 10
		if x == 1 {
			return one
		} else if x >= 2 && x <= 4 {
			return two
		} else {
			return five
		}
	} else {
		return five
	}
}

func inArray(haystack []string, needle string) bool  {
	for _,v:=range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func trimAndLower(str string) string {
	return strings.TrimSpace( strings.ToLower(str) )
}

func postJsonRequest(url string, request string, v interface{}) ([]byte, error) {
	t := &http.Client{Timeout: time.Second * 300}
	var r []byte
	resp, err :=t.Post(url, "application/json", bytes.NewBuffer([]byte(request)))
	if err != nil {
		return r, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return r, err
	} else {
		return data, nil
	}
}