package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type IpResult struct {
	source string
	IPv4   string
	IPv6   string
}

type IpManager struct {
	IPv4 string
	IPv6 string

	ipv4Reg  *regexp.Regexp
	ipv6Reg  *regexp.Regexp
	Timeout  float64
	workChan chan IpResult
	repeat   int
}

func NewIpManager() *IpManager {
	return &IpManager{
		IPv4:    "",
		IPv6:    "",
		Timeout: 5,
		ipv4Reg: regexp.MustCompile("^((25[0-5]|2[0-4]\\d|[01]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)$"),
		ipv6Reg: regexp.MustCompile("^\\s*((([0-9A-Fa-f]{1,4}:){7}(([0-9A-Fa-f]{1,4})|:))|(([0-9A-Fa-f]{1,4}:){6}(:|((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})|(:[0-9A-Fa-f]{1,4})))|(([0-9A-Fa-f]{1,4}:){5}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){4}(:[0-9A-Fa-f]{1,4}){0,1}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){3}(:[0-9A-Fa-f]{1,4}){0,2}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){2}(:[0-9A-Fa-f]{1,4}){0,3}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:)(:[0-9A-Fa-f]{1,4}){0,4}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(:(:[0-9A-Fa-f]{1,4}){0,5}((:((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|" +
			"(((25[0-5]|2[0-4]\\d|[01]?\\d{1,2})(\\.(25[0-5]|2[0-4]\\d|[01]?\\d{1,2})){3})))(%.+)?\\s*$"),
		workChan: make(chan IpResult),
		repeat:   3,
	}
}

func (ip *IpManager) SetIP(ipv4 string) {
	ip.SetIPv4(ipv4)

}
func (ip *IpManager) SetIPv4(ipv4 string) {
	IPv4IsLegal := ip.ipv4Reg.MatchString(ipv4)
	if IPv4IsLegal {
		ip.IPv4 = ipv4
	} else {
		log.Println("IPv4地址错误: ", ipv4)
	}
}

func (ip *IpManager) SetIPv6(ipv6 string) {
	IPv6IsLegal := ip.ipv6Reg.MatchString(ipv6)
	if IPv6IsLegal {
		ip.IPv6 = ipv6
	} else {
		log.Println("非法IPv6地址: ", ipv6)
	}
}

func (ip *IpManager) GetIP() string {
	return ip.IPv4
}
func (ip *IpManager) GetIPv4() string {
	return ip.IPv4
}
func (ip *IpManager) GetIPv6() string {
	return ip.IPv6
}

func (ip *IpManager) GetPublicIpAddress() {
	for times := 0; times < ip.repeat; times++ {
		go ip.GetIpByWhatIsMyIpAddress(ip.workChan)
		go ip.GetIpByTestIpv6Web(ip.workChan)
	}

	for times := 0; times < ip.repeat*2; times++ {
		result := <-ip.workChan
		//debug---------------------------------------------------
		if result.source == "test-ipv6.com" {
			test_ipv6Result := result
			log.Printf("test-ipv6.com: ipv4:[%s] ipv6[%s] \n", test_ipv6Result.IPv4, test_ipv6Result.IPv6)
		} else {
			whatismyipaddressResult := result
			log.Printf("whatismyipaddressResult.com: ipv4:[%s] ipv6:[%s] \n", whatismyipaddressResult.IPv4, whatismyipaddressResult.IPv6)
		}
		//---------------------------------------------------

		if result.IPv4 != "" {
			if ip.IPv4 != "" {
				if ip.IPv4 != result.IPv4 {
					log.Println("检测IPv4不一致", result.IPv4, ip.IPv4)
				}
			} else {
				ip.SetIPv4(result.IPv4)
			}
		}
		if result.IPv6 != "" {
			if ip.IPv6 != "" {
				if ip.IPv6 != result.IPv6 {
					log.Println("检测IPv6不一致", result.IPv6, ip.IPv6)
				}
			} else {
				ip.SetIPv6(result.IPv6)
			}
		}
	}

	//defer close(ip.workChan)
	return
}

func (ip *IpManager) GetIpByTestIpv6Web(workChan chan IpResult) {
	ipResult := IpResult{source: "test-ipv6.com"}
	Ipv4TestUrl := "http://ipv4.test-ipv6.com/ip/"
	Ipv6TestUrl := "http://ipv6.test-ipv6.com/ip/"
	HttpClient := &http.Client{Timeout: time.Duration(ip.Timeout) * time.Second}
	Ipv4Resp, Ipv4Err := HttpClient.Get(Ipv4TestUrl)
	Ipv6Resp, Ipv6Err := HttpClient.Get(Ipv6TestUrl)

	type callback struct {
		Ip      string `json:"ip"`
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		Via     string `json:"via"`
		Padding string `json:"padding"`
	}
	if Ipv4Err != nil {
		log.Printf("HTTP请求不可达[%s]: %s\n", ipResult.source, Ipv4Err.Error())
	} else {
		body, _ := ioutil.ReadAll(Ipv4Resp.Body)
		bodyString := string(body)
		bodyJson := bodyString[9 : len(bodyString)-2]
		var result callback
		if err := json.Unmarshal([]byte(bodyJson), &result); err == nil {
			ipv4 := strings.Split(result.Ip, ",")[0]
			ipResult.IPv4 = ipv4
		} else {
			log.Println(err)
		}

		defer Ipv4Resp.Body.Close()

	}

	if Ipv6Err != nil {
		log.Printf("HTTP请求不可达[%s]: %s\n", ipResult.source, Ipv6Err.Error())

	} else {
		body, _ := ioutil.ReadAll(Ipv6Resp.Body)
		bodyString := string(body)
		bodyJson := bodyString[9 : len(bodyString)-2]
		var result callback
		if err := json.Unmarshal([]byte(bodyJson), &result); err == nil {
			ipv6 := strings.Split(result.Ip, ",")[0]
			ipResult.IPv6 = ipv6
		} else {
			log.Println(err)
		}

		defer Ipv6Resp.Body.Close()
	}

	//接受结果
	workChan <- ipResult
	return
}

func (ip *IpManager) GetIpByWhatIsMyIpAddress(workChan chan IpResult) {
	ipResult := IpResult{source: "whatismyipaddress.com"}
	ipv4Url := "http://ipv4bot.whatismyipaddress.com"
	ipv6Url := "http://ipv6bot.whatismyipaddress.com"
	HttpClient := &http.Client{Timeout: time.Duration(ip.Timeout) * time.Second}
	Ipv4Resp, Ipv4Err := HttpClient.Get(ipv4Url)
	Ipv6Resp, Ipv6Err := HttpClient.Get(ipv6Url)
	if Ipv4Err != nil {
		log.Printf("HTTP请求不可达[%s]: %s\n", ipResult.source, Ipv4Err.Error())
	} else {
		body, _ := ioutil.ReadAll(Ipv4Resp.Body)
		ipResult.IPv4 = string(body)
		_ = Ipv4Resp.Body.Close()
	}
	if Ipv6Err != nil {
		log.Printf("HTTP请求不可达[%s]: %s\n", ipResult.source, Ipv6Err.Error())
	} else {
		body, _ := ioutil.ReadAll(Ipv6Resp.Body)
		ipResult.IPv6 = string(body)
		Ipv6Resp.Body.Close()
	}
	workChan <- ipResult
	return
}

// 从网口获取IPv6信息
func (ip *IpManager) GetIpByNICAddress(workChan chan IpResult) {
	ipResult := IpResult{source: "net.Interfaces()"}
	inters, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, inter := range inters {
		// 判断网卡是否开启，过滤本地环回接口
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			// 获取网卡下所有的地址
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					//判断是否存在IPV6 IP 如果没有过滤
					if ipnet.IP.To4() != nil {
						ipResult.IPv4 = ipnet.IP.String()
					}
				}
			}
		}
	}
	workChan <- ipResult
	return
}
