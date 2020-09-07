package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type CloudflareAPI struct {
	CloudflareConf CloudflareConf
	RootSite string
}

type CloudflareConf struct {
	Email      string `json:"Email"`
	ApiKey     string `json:"API_key"`
	ZoneId     string `json:"Zone_ID"`
}


// 记录DNS查询结果
type QueueDNSResult struct {
	Result []struct {
		ID        string `json:"id"`
		ZoneID    string `json:"zone_id"`
		ZoneName  string `json:"zone_name"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Content   string `json:"content"`
		Proxiable bool   `json:"proxiable"`
		Proxied   bool   `json:"proxied"`
		TTL       int    `json:"ttl"`
		Locked    bool   `json:"locked"`
		Meta      struct {
			AutoAdded           bool   `json:"auto_added"`
			ManagedByApps       bool   `json:"managed_by_apps"`
			ManagedByArgoTunnel bool   `json:"managed_by_argo_tunnel"`
			Source              string `json:"source"`
		} `json:"meta"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
		TotalPages int `json:"total_pages"`
	} `json:"result_info"`
}

// 记录DNS更新结果
type UpdateCreateDNSResult struct {
	Result struct {
		ID string `json:"id"`
		ZoneID string `json:"zone_id"`
		ZoneName string `json:"zone_name"`
		Name string `json:"name"`
		Type string `json:"type"`
		Content string `json:"content"`
		Proxiable bool `json:"proxiable"`
		Proxied bool `json:"proxied"`
		TTL int `json:"ttl"`
		Locked bool `json:"locked"`
		Meta struct {
			AutoAdded bool `json:"auto_added"`
			ManagedByApps bool `json:"managed_by_apps"`
			ManagedByArgoTunnel bool `json:"managed_by_argo_tunnel"`
			Source string `json:"source"`
		} `json:"meta"`
		CreatedOn time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

// HTTp数据包
type HttpData struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Content string `json:"content"`
	TTL int `json:"ttl"`
	Proxied bool `json:"proxied"`
}

func NewCloudflareAPI() *CloudflareAPI {
	configJsonFile, err := os.Open("CloudflareConf.json")
	if err != nil {
		panic(err)
	}
	byteValue, _ := ioutil.ReadAll(configJsonFile)
	var cloudflareConf CloudflareConf
	if err := json.Unmarshal(byteValue, &cloudflareConf); err == nil {
		return &CloudflareAPI{
			CloudflareConf: cloudflareConf,
			RootSite: "https://api.cloudflare.com/client/v4",
		}
	}else {
		panic(err)
	}

}

func (api *CloudflareAPI)ListDNSRecords(DNStype string, Domain string) QueueDNSResult{
	var queueDNSResult QueueDNSResult
	client := &http.Client{}
	listUrl := fmt.Sprintf("%s/zones/%s/dns_records", api.RootSite, api.CloudflareConf.ZoneId)
	uri, _ := url.Parse(listUrl)
	queryValues := url.Values{}
	queryValues.Add("type", DNStype)
	queryValues.Add("name", Domain)
	uri.RawQuery = queryValues.Encode()

	req, _ := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("X-Auth-Email", api.CloudflareConf.Email)
	req.Header.Set("X-Auth-Key", api.CloudflareConf.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return queueDNSResult
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &queueDNSResult)
	if  err != nil {
		panic(err)
	}
	return queueDNSResult

}

func (api *CloudflareAPI)UpdateDNSRecord(IpProtocol string, IpAddress string, RecordId, IpDpmain string)  UpdateCreateDNSResult{
	var updateDNSResult UpdateCreateDNSResult
	updateData := &HttpData{
		Name: IpDpmain,
		Content: IpAddress,
		TTL: 1,
		Proxied: false,
	}
	if IpProtocol == "ipv4" {
		updateData.Type = "A"
	}else if IpProtocol == "ipv6" {
		updateData.Type = "AAAA"
	}
	updateDataJson, err := json.Marshal(updateData)
	if err != nil {
		log.Println(err)
	}
	client := &http.Client{}
	updateUrl := fmt.Sprintf("%s/zones/%s/dns_records/%s", api.RootSite, api.CloudflareConf.ZoneId, RecordId)
	req, _ := http.NewRequest("PUT", updateUrl, bytes.NewReader(updateDataJson))
	req.Header.Set("X-Auth-Email", api.CloudflareConf.Email)
	req.Header.Set("X-Auth-Key", api.CloudflareConf.ApiKey)
	req.Header.Set("Content-Type", "application/json")


	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &updateDNSResult)
	if  err != nil {
		panic(err)
	}
	return updateDNSResult
}

func (api *CloudflareAPI)CreateDNSRecord(IpProtocol string, IpAddress string, IpDpmain string) UpdateCreateDNSResult{
	var createDNSResult UpdateCreateDNSResult

	createData := &HttpData{
		Name: IpDpmain,
		Content: IpAddress,
		TTL: 1,
		Proxied: false,
	}
	if IpProtocol == "ipv4" {
		createData.Type = "A"
	}else if IpProtocol == "ipv6" {
		createData.Type = "AAAA"
	}
	createDataJson, err := json.Marshal(createData)
	if err != nil {
		log.Println(err)
	}
	client := &http.Client{}
	createUrl := fmt.Sprintf("%s/zones/%s/dns_records", api.RootSite, api.CloudflareConf.ZoneId)
	req, _ := http.NewRequest("POST", createUrl, bytes.NewReader(createDataJson))
	req.Header.Set("X-Auth-Email", api.CloudflareConf.Email)
	req.Header.Set("X-Auth-Key", api.CloudflareConf.ApiKey)
	req.Header.Set("Content-Type", "application/json")


	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return createDNSResult
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &createDNSResult)
	if  err != nil {
		panic(err)
	}
	return createDNSResult
}
