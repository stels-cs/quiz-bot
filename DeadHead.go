package main

import (
	"net/http"
	"fmt"
	"github.com/stels-cs/quiz-bot/Vk"
	"strings"
	"time"
	"io/ioutil"
	"encoding/base64"
	"strconv"
)

const captchaFromTemplate = `
<h1 style="color:#800">Captcha</h1>

<img src="%s"><br/>
<form action="/captcha/" method="POST">
<input type="text" name="code" placeholder="code"/>
<input type="submit"/>
</form>
`

const saveCaptchaResponse = `
<h1>Ok, please wait...</h1>
<script>
setTimeout( function () { window.location = '/'; }, 5000 );
</script>
`

type AntiCpatchaRes struct {
	ErrorId int `json:"errorId"`
	TaskId  int `json:"taskId"`
}

type AntiCpatchaTaskResult struct {
	ErrorId int `json:"errorId"`
	Status string `json:"status"`
	Solution struct{
		Text string `json:"text"`
	} `json:"solution"`
}

type DeadHand struct {
	Captcha        string
	server         *http.Server
	api            *Vk.Api
	bot            *Bot
	antiCaptchaKey string
}

func GetDeadHand(api *Vk.Api, bot *Bot) *DeadHand {
	me := DeadHand{api: api, bot: bot}
	me.antiCaptchaKey = "d22724aa79019444391b4c718092cc53"
	api.SetCaptchaListener(&me)
	return &me
}

func (dh *DeadHand) recover(key string) {
	if dh.Captcha != "" && key != "" {
		dh.Captcha = ""
		dh.api.SetCaptchaKey(key)
		dh.bot.OnCaptchaRecover()
	}
}

func (dh *DeadHand) viewHandler(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("code")
	dh.recover(key)
	fmt.Fprint(w, saveCaptchaResponse)
}

func (dh *DeadHand) mainHandler(w http.ResponseWriter, r *http.Request) {
	if dh.Captcha == "" {
		fmt.Fprint(w, "<h1 style='color:#080'>Work normal!</h1>")
	} else {
		fmt.Fprintf(w, captchaFromTemplate, dh.Captcha)
	}
}

func (dh *DeadHand) antiCaptcha(img string) {
	print("Recatpcha " + img + "\n")
	t := &http.Client{Timeout: time.Second * 300}
	resp, err := t.Get(img)
	if err != nil {
		print(err)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		print(err)
		return
	}
	image := base64.StdEncoding.EncodeToString(data)

	request := `{"clientKey":"%apikey%","task":{"type":"ImageToTextTask","body":"%body%","phrase":false},"languagePool":"rn"}`
	request = strings.Replace(request, "%apikey%", dh.antiCaptchaKey, 1)
	request = strings.Replace(request, "%body%", image, 1)
	ac := AntiCpatchaRes{}
	data, err = postJsonRequest("https://api.anti-captcha.com/createTask", request, &ac)
	if err != nil {
		print(err.Error() + "\n")
		return
	}
	if ac.TaskId == 0 {
		print(string(data) + "\n")
		return
	}

	time.Sleep(5 * time.Second)
	i := 0
	checkRequest := `{"clientKey":"%apikey%","taskId":%taskId%}`
	checkRequest = strings.Replace(checkRequest, "%apikey%", dh.antiCaptchaKey, 1)
	checkRequest = strings.Replace(checkRequest, "%taskId%", strconv.Itoa(ac.TaskId), 1)
	for i < 20 {
		i++
		acr := AntiCpatchaTaskResult{}
		data, err = postJsonRequest("https://api.anti-captcha.com/getTaskResult", checkRequest, &acr)
		if acr.ErrorId != 0 {
			if ac.TaskId == 0 {
				print(string(data) + "\n")
				return
			}
		}
		if acr.Status == "ready" {
			print("Ready! " + acr.Solution.Text + "\n")
			dh.recover(acr.Solution.Text)
			return
		} else {
			print("Anti captcha " + acr.Status + "\n")
		}

		time.Sleep(3 * time.Second)
	}
	print("Captcha not discovered\n")
}

func (dh *DeadHand) OnCaptcha(img string) {
	if dh.Captcha == "" {
		go dh.antiCaptcha(img)
	}
	dh.Captcha = strings.TrimSpace(img)

}

func (dh *DeadHand) GetName() string {
	return "DeadHand"
}

func (dh *DeadHand) Start() error {
	mux := &http.ServeMux{}
	mux.HandleFunc("/captcha/", dh.viewHandler)
	mux.HandleFunc("/", dh.mainHandler)
	dh.server = &http.Server{Addr: ":8088", Handler: mux}
	return dh.server.ListenAndServe()
}

func (dh *DeadHand) Stop() {
	dh.server.Close()
}
