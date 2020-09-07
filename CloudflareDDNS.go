package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DomainsConfig struct {
	UpdateIpv6 bool   `json:"Update_IPv6"`
	Ipv6Domain string `json:"IPv6_domain"`
	UpdateIpv4 bool   `json:"Update_IPv4"`
	Ipv4Domain string `json:"IPv4_domain"`
}

func main() {


	//创建IP管理器
	ipManager := NewIpManager()
	//创建CloudflareAPI
	api := NewCloudflareAPI()

	//周期性任务函数
	CyclicTask :=func(){
		//获取公网IP
		ipManager.GetPublicIpAddress()

		//读取本地绑定域名参数json
		configJsonFile, err := os.Open("domains.json")
		if err != nil {
			panic(err)
		}
		byteValue, _ := ioutil.ReadAll(configJsonFile)
		var domainsConfig DomainsConfig
		if err := json.Unmarshal(byteValue, &domainsConfig); err == nil {

			if ipManager.IPv4 == "" && ipManager.IPv6 == "" {
				log.Println("未获取到有效IPv4或Ipv6地址，请检查网络")
				return
			}
			log.Println("获取到公网IPv4: ", ipManager.IPv4, "\tIPv6: ", ipManager.IPv6)
			//添加或修改DNS记录
			if domainsConfig.UpdateIpv4 == true {
				resultA := api.ListDNSRecords("A", domainsConfig.Ipv4Domain)
				if len(resultA.Result) == 0 {
					//没有查询到记录
					createResult := api.CreateDNSRecord("ipv4", ipManager.IPv4, domainsConfig.Ipv4Domain)
					if createResult.Success {
						log.Printf("创建DNS成功 \n\tID: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", createResult.Result.ID, createResult.Result.Type, createResult.Result.Name, createResult.Result.Content)
					} else {
						log.Printf("创建DNS失败 \n\tMessage: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", createResult.Errors, createResult.Result.Type, createResult.Result.Name, createResult.Result.Content)
					}
				} else {
					for _, queueResult := range resultA.Result {
						updateResult := api.UpdateDNSRecord("ipv4", ipManager.IPv4, queueResult.ID, domainsConfig.Ipv4Domain)
						if updateResult.Success {
							log.Printf("更新DNS成功 \n\tID: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s",
								updateResult.Result.ID, updateResult.Result.Type, updateResult.Result.Name, updateResult.Result.Content)
						} else {
							log.Printf("更新DNS失败 \n\tError: %s \n\tMessage: %s \n\tDomain: %s \n\tIP: %s",
								updateResult.Errors, updateResult.Errors, updateResult.Result.Name, updateResult.Result.Content)
						}
					}

				}

			}
			if domainsConfig.UpdateIpv6 == true {
				resultAAAA := api.ListDNSRecords("AAAA", domainsConfig.Ipv6Domain)
				if len(resultAAAA.Result) == 0 {
					//没有查询到记录
					createResult := api.CreateDNSRecord("ipv6", ipManager.IPv6, domainsConfig.Ipv6Domain)
					if createResult.Success {
						log.Printf("创建DNS成功 \n\tID: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", createResult.Result.ID, createResult.Result.Type, createResult.Result.Name, createResult.Result.Content)
					} else {
						log.Printf("创建DNS失败 \n\tMessage: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", createResult.Errors, createResult.Result.Type, createResult.Result.Name, createResult.Result.Content)
					}
				} else {
					for _, queueResult := range resultAAAA.Result {
						updateResult := api.UpdateDNSRecord("ipv6", ipManager.IPv6, queueResult.ID, domainsConfig.Ipv6Domain)
						if updateResult.Success {
							log.Printf("更新DNS成功 \n\tID: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", updateResult.Result.ID, updateResult.Result.Type, updateResult.Result.Name, updateResult.Result.Content)
						} else {
							log.Printf("更新DNS失败 \n\tMessage: %s \n\tType: %s \n\tDomain: %s \n\tIP: %s", updateResult.Errors, updateResult.Result.Type, updateResult.Result.Name, updateResult.Result.Content)
						}

					}

				}
			}
		}
	}
	tickSecondTime := 10*60//定时任务second
	var timesChannel chan int //定义运行计数ch
	ticker := time.NewTicker(time.Second * time.Duration(tickSecondTime))

	go func() {
		CyclicTask()
		for range ticker.C {
			CyclicTask()
		}
		timesChannel <- 1
	}()
	<-timesChannel
	return
}
