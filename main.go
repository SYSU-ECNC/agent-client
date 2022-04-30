package main

import (
	"agent-client/config"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/bitly/go-simplejson"
)

var agentInfo config.AgentInfo

func init() {
	agentInfo = config.ReadConfig()
	for !register() {
		// 一直注册直至注册成功
		config.WriteAgentInfoInConfig()
		agentInfo = config.ReadConfig()
	}
}

func register() bool {
	url := `http://` + agentInfo.MasterIP + `:` + agentInfo.MasterPort + `/agent/new`
	data := `{
		"name":"` + agentInfo.Name + `",
		"token":"` + agentInfo.Token + `"
		}`
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("register resp.StatusCode: ", resp.StatusCode)
		return false
	}

	return true
}

var taskID = make(chan int, 1)
var command = make(chan string, 1)
var result = make(chan string, 1)

func jsonObj2Value(respBody io.Reader) {
	body, _ := ioutil.ReadAll(respBody)
	respData, _ := simplejson.NewJson(body)
	id, _ := respData.Get("task_id").Int()
	cmd, _ := respData.Get("command").String()

	if id == 0 || cmd == "" {
		// 此时该agent不存在未完成的任务
		return
	}

	taskID <- id
	command <- cmd
}

func getTask() {
	url := `http://` + agentInfo.MasterIP + `:` + agentInfo.MasterPort + `/agent/task?agent_name=` + agentInfo.Name

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Authorization", agentInfo.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("getTask resp.StatusCode: ", resp.StatusCode)
		return
	}
	jsonObj2Value(resp.Body)
}

func gbk2UTF8(src string) string {
	// 使用GBK编码是因为agent-client运行环境有可能是中文的
	srcCoder := mahonia.NewDecoder("gbk")
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder("utf-8")
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)

	return string(cdata)
}

func execTask() {

	// log.Println("exec task")

	cmd := strings.Fields(<-command)
	out, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		log.Fatalln(err)
	}

	//对运行结果进行处理
	execResult := ""
	for _, v := range strings.Fields(gbk2UTF8(string(out))) {
		execResult = execResult + v + " "
	}

	// 输出执行结果
	log.Println(execResult)

	result <- execResult

}

func reportResult() {

	url := `http://` + agentInfo.MasterIP + `:` + agentInfo.MasterPort + `/agent/task`

	data := `{
		"name":"` + agentInfo.Name + `",
		"task_id":` + fmt.Sprint(<-taskID) + `,
		"result":"` + <-result + `"
		}`

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(data))

	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Authorization", agentInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("reportResult resp.StatusCode: ", resp.StatusCode)
	}

	// log.Println("report task")

}

func main() {
	for {
		go getTask()
		go execTask()
		go reportResult()
		time.Sleep(1 * time.Second)
	}
}
