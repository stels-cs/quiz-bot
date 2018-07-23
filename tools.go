package main

import (
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
	return ff(v != "", v, def)
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

func inArray(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func trimAndLower(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
}

func postJsonRequest(url string, request string, v interface{}) ([]byte, error) {
	t := &http.Client{Timeout: time.Second * 300}
	var r []byte
	resp, err := t.Post(url, "application/json", bytes.NewBuffer([]byte(request)))
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

func trimToken(token string) string {
	if len(token) > 10 {
		left := string([]rune(token)[0 : 5])
		right := string([]rune(token)[len(token)-6 : len(token)])
		return left + "..." + right
	} else {
		return token
	}
}
