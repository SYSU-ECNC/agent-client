package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AgentInfo struct {
	Name       string
	MasterIP   string
	MasterPort string
	Token      string
}

var agentInfo AgentInfo

func ScanAgentInfoFromConsole() {
	fmt.Println("Please enter the agent name: ")
	fmt.Scan(&agentInfo.Name)
	fmt.Println("Please enter the master's ip: ")
	fmt.Scan(&agentInfo.MasterIP)
	fmt.Println("Please enter the master's port: ")
	fmt.Scan(&agentInfo.MasterPort)
	fmt.Println("Please enter the  registration token: ")
	fmt.Scan(&agentInfo.Token)
}

func WriteAgentInfoInConfig() {
	ScanAgentInfoFromConsole()
	viper.Set("name", agentInfo.Name)
	viper.Set("masterIP", agentInfo.MasterIP)
	viper.Set("masterPort", agentInfo.MasterPort)
	viper.Set("token", agentInfo.Token)
	viper.WriteConfig()
}

func init() {
	viper.AddConfigPath("./config/")
	// viper.SetConfigName("config")
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		WriteAgentInfoInConfig()
		return
	}

	agentInfo.Name = viper.Get("name").(string)
	agentInfo.MasterIP = viper.Get("masterIP").(string)
	agentInfo.MasterPort = viper.Get("masterPort").(string)
	agentInfo.Token = viper.Get("token").(string)
}

func ReadConfig() AgentInfo {
	return agentInfo
}
