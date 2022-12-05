package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	jar, _ := cookiejar.New(nil)
	client := http.Client{
		Jar: jar,
	}
	phone := os.Getenv("phone")
	login(client, phone)
	//for i := 0; i < 4; i++ {
	//	answerQuestion(client)
	//}
	pushMessage(client)
}

func pushMessage(client http.Client) {
	mainReq, err := http.NewRequest(http.MethodGet, "https://exam.gog.cn/ExamDSJS3/main", nil)
	if err != nil {
		panic(err)
	}
	mainResp, err := client.Do(mainReq)
	if err != nil {
		panic(err)
	}
	defer mainResp.Body.Close()
	mainResult, err := goquery.NewDocumentFromReader(mainResp.Body)
	text := mainResult.Find(".score-today").Text()
	fmt.Println(text)
	markdown := fmt.Sprintf("当日答题已完成，当天分数%v\n\n详情见[https://exam.gog.cn/ExamDSJS3](https://exam.gog.cn/ExamDSJS3)\n", text)
	sendKey := os.Getenv("sendKey")
	var param = make(map[string]string)
	param["title"] = "答题结果"
	param["desp"] = markdown
	param["channel"] = "9"
	paramJson, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	if len(sendKey) > 0 {
		msgRequest, err := http.NewRequest(http.MethodPost, "https://sctapi.ftqq.com/"+sendKey+".send", bytes.NewReader(paramJson))
		if err != nil {
			panic(err)
		}
		msgRequest.Header.Add("Content-Type", "application/json")
		//client.Do(msgRequest)
	}
}

func answerQuestion(client http.Client) {
	check(client)
	pageResult := getExamPage(client)
	defer pageResult.Body.Close()
	doc, err := goquery.NewDocumentFromReader(pageResult.Body)
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UnixNano())
	second := rand.Intn(25-10+1) + 10
	time.Sleep(time.Second * time.Duration(second))
	var submitList = make([]Submit, 10)
	doc.Find(".item").Each(func(i int, selection *goquery.Selection) {
		subjectTypeIdStr, exists := selection.Attr("data-typeid")
		subjectTypeId, err := strconv.Atoi(subjectTypeIdStr)
		if err != nil {
			fmt.Println(err)
		}
		subjectId, exists := selection.Attr("data-subjectid")
		optionAnswer, exists := selection.Attr("data-answer")
		if subjectTypeId == 1 {
			submit := Submit{
				SubjectId:     subjectId,
				SubjectTypeId: subjectTypeId,
				OptionAnswer:  optionAnswer,
				WrongStatus:   0,
			}
			submitList[i] = submit
		} else {
			submit := Submit{
				SubjectId:     subjectId,
				SubjectTypeId: subjectTypeId,
				OptionAnswer:  optionAnswer,
			}
			submitList[i] = submit
		}
		if exists {
			fmt.Printf("typeId=%v,subjectId=%v,answer=%v\n", subjectTypeId, subjectId, optionAnswer)
		}
	})
	marshal, err := json.Marshal(submitList)
	if err != nil {
		panic(err)
	}
	param := "paper=" + string(marshal)
	fmt.Println("参数=>" + param)
	submitReq, err := http.NewRequest(http.MethodPost, "https://exam.gog.cn/ExamDSJS3/Main/Submit", strings.NewReader(param))
	if err != nil {
		panic(err)
	}
	submitReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	submitResp, err := client.Do(submitReq)
	if err != nil {
		panic(err)
	}
	defer submitResp.Body.Close()
	all, err := ioutil.ReadAll(submitResp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("result is %v\n", string(all))
}

func getExamPage(client http.Client) *http.Response {
	request, err := http.NewRequest(http.MethodGet, "https://exam.gog.cn/ExamDSJS3/main/exam", nil)
	if err != nil {
		panic(err)
	}
	do, err := client.Do(request)
	return do
}

func check(client http.Client) {
	request, err := http.NewRequest(http.MethodPost, "https://exam.gog.cn/ExamDSJS3/Main/CheckExam", nil)
	if err != nil {
		panic(err)
	}
	do, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer do.Body.Close()
	all, err := ioutil.ReadAll(do.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(all))
}

func login(client http.Client, phone string) {
	param := make(map[string]string)
	param["Phone"] = phone
	marshal, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshal))
	request, err := http.NewRequest(http.MethodPost, "https://exam.gog.cn/ExamDSJS3/Home/Login", bytes.NewReader(marshal))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	do, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	all, err := ioutil.ReadAll(do.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(all))
}

type Submit struct {
	SubjectId     string `json:"subjectId"`
	SubjectTypeId int    `json:"subjectTypeId"`
	OptionAnswer  string `json:"optionAnswer"`
	WrongStatus   int    `json:"wrongStatus"`
}
