package main

import(
    "os"
    "log"
    "time"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "encoding/json"
)

var config *Config
var domainInfos []*DomainInfo

type Config struct {
    DomainNames []string   `json:"DomainNames"`
    DnspodEmail  string    `json:"DnspodEmail"`
    DnspodPasswd string    `json:"DnspodPasswd"`
}

type DnsPodResponseStatus struct {
    Code        string  `json:"code"`
    Message     string  `json:"message"`
    CreatedAt   string  `json:"created_at"`
}

type DomainInfo struct {
    Id string
    Name string
    RecordId string
    CurrentIp string
}

func main() {
    chs := make([]chan int, len(domainInfos))
    for i, domainInfo := range domainInfos {
        go loopUpdate(domainInfo, chs[i])
    }

    for _, ch := range chs {
        <-ch
    }
    log.Println("ddns quit")
}

func loopUpdate(domainInfo *DomainInfo, ch chan int) {
    ticker := time.NewTicker(5 * time.Second)
    for {
        select {
        case <-ticker.C:
            localIp := getLocalIp()
            if len(localIp) < 1 {
                continue
            }

            if localIp != domainInfo.CurrentIp {
                log.Println("start update", domainInfo.Name, domainInfo.CurrentIp, "to", localIp)
                if ok := updateDomainDNS(domainInfo, localIp); ok {
                    log.Println("success to update ", domainInfo.Name, " ip to", domainInfo.CurrentIp)
                } else {
                    log.Println("failed to update ", domainInfo.Name, " ip to", domainInfo.CurrentIp)
                }
            } else {
                log.Println("no need to update", domainInfo.Name)
            }
        }
    }
    ch <- 1
}

func init() {
    if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
        log.Fatalln("config.json not found")
    }
    content, err := ioutil.ReadFile("./config.json")
    if err != nil {
        log.Fatalln("failed to read config.json ", err.Error())
    }
    config = &Config{}
    if err := json.Unmarshal(content, config); err != nil {
        log.Fatalln("failed to parse config.json ", err.Error())
    }

    for _, domainName := range config.DomainNames {
        domainInfos = append(domainInfos, initDomainInfo(domainName))
    }

    if len(domainInfos) < 1 {
        log.Fatalln("cannot find any domain info")
    }
}

func getLocalIp() (ip string) {
    ipUrl := "http://agideo.com/ip"
    res, err := http.Get(ipUrl)
    if err != nil {
        log.Println("cannot get local ip via ", ipUrl, err.Error())
        return
    }
    defer res.Body.Close()
    cip, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        log.Println("read err, cannot get local ip via ", ipUrl, err.Error())
        return
    }

    return string(cip)
}

func initDomainInfo(domainName string) (info *DomainInfo) {
    domainId := getDomainId(domainName)
    if len(domainId) < 1 {
        log.Fatalln("failed to get domain id", domainName)
        return
    }
    recordId, recordIp := getRecordIdIp(domainId, domainName)
    if len(recordId) < 1 || len(recordIp) < 1 {
        log.Fatalln("failed to get record id ", domainName)
        return
    }
    info = &DomainInfo{Id: domainId, Name: domainName, RecordId: recordId, CurrentIp: recordIp}
    return
}

func updateDomainDNS(domainInfo *DomainInfo, ip string) (ok bool) {
    updateUrl := "https://dnsapi.cn/Record.Ddns"
    parms     := make(url.Values, 0)

    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain_id", domainInfo.Id)
    parms.Add("record_id", domainInfo.RecordId)
    parms.Add("sub_domain", strings.Split(domainInfo.Name, ".")[0])
    parms.Add("value", ip)
    parms.Add("record_line", "默认")
    parms.Add("format", "json")

    res, err := http.PostForm(updateUrl, parms)
    if err != nil {
        log.Println("failed update domain dns", updateUrl)
        return
    }
    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        log.Println("cannot get update damian info ", domainInfo.Name, err1.Error())
        return
    }

    type Response struct {
        Status DnsPodResponseStatus
    }
    var response Response
    if err := json.Unmarshal(body, &response); err != nil {
        log.Println("failed to parse update domain info ", err.Error())
        return
    }

    if response.Status.Code != "1" {
        log.Println("failed to update domain dns, status code:", response.Status.Code, response.Status.Message)
        return
    }

    domainInfo.CurrentIp = ip
    ok = true
    return
}

func getDomainId(domainName string) (domainId string) {
    domainUrl := "https://dnsapi.cn/Domain.Info"
    parms     := make(url.Values, 0)
    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain", strings.Join(strings.Split(domainName, ".")[1:], "."))
    parms.Add("format", "json")

    res, err := http.PostForm(domainUrl, parms)
    if err != nil {
        log.Println("cannot get damian id ", domainName, err.Error())
        return
    }

    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        log.Println("cannot get damian id ", domainName, err1.Error())
        return
    }

    type DomainResponse struct {
        Status DnsPodResponseStatus
        Domain struct {
            Id string
            Name string
        }
    }
    var dr DomainResponse
    if err := json.Unmarshal(body, &dr); err != nil {
        log.Println("failed to parse domain info ", err.Error())
        return
    }

    if dr.Status.Code != "1" {
        log.Println("domain info invalid, status code:", dr.Status.Code, dr.Status.Message)
        return
    }

    return dr.Domain.Id
}

func getRecordIdIp(domainId string, domainName string) (recordId, recordIp string) {
    recordUrl := "https://dnsapi.cn/Record.List"
    parms     := make(url.Values, 0)

    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain_id", domainId)
    parms.Add("sub_domain", strings.Split(domainName, ".")[0])
    parms.Add("format", "json")

    res, err := http.PostForm(recordUrl, parms)
    if err != nil {
        log.Println("cannot get record id via ", recordUrl, err.Error())
        return
    }

    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        log.Println("cannot get record id via ", recordUrl, err1.Error())
        return
    }

    type RecordResponse struct {
        Status  DnsPodResponseStatus
        Records []struct {
            Id      string
            Name    string
            Type    string
            Value   string
            Status  string
            Enabled string
        }
    }
    var rr RecordResponse
    if err := json.Unmarshal(body, &rr); err != nil {
        log.Println("failed to parse record info, ", err.Error())
        return
    }
    if rr.Status.Code == "10" {
        recordId, recordIp, _ = createRecord(domainId, domainName)
        return
    }
    if rr.Status.Code != "1" {
        log.Println("record info invalid, status code: ", rr.Status.Code, rr.Status.Message)
        return
    }

    if len(rr.Records) > 0 {
        recordId = rr.Records[0].Id
        recordIp = rr.Records[0].Value
    }
    return
}

func createRecord(domainId, domainName string) (recordId, recordIp string, ok bool) {
    recordUrl := "https://dnsapi.cn/Record.Create"
    parms     := make(url.Values, 0)
    ip := getLocalIp()

    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain_id", domainId)
    parms.Add("sub_domain", strings.Split(domainName, ".")[0])
    parms.Add("record_type", "A")
    parms.Add("record_line", "默认")
    parms.Add("value", ip)
    parms.Add("mx", "10")
    parms.Add("format", "json")

    res, err := http.PostForm(recordUrl, parms)
    if err != nil {
        log.Println("cannot create record", domainName, err.Error())
        return
    }

    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        log.Println("failed to parse create record", domainName, err1.Error())
        return
    }

    type RecordResponse struct {
        Status  DnsPodResponseStatus
        Record struct {
            Id      string
            Name    string
            Status  string
        }
    }
    var rr RecordResponse
    if err := json.Unmarshal(body, &rr); err != nil {
        log.Println("failed to parse create record info, ", err.Error())
        return
    }
    if rr.Status.Code != "1" {
        log.Println("create record info invalid, status code: ", rr.Status.Code, rr.Status.Message)
        return
    }

    log.Println("success to create record", domainName, ip)
    return rr.Record.Id, ip, true
}
