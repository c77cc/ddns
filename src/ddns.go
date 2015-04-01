package main

import (
    "fmt"
    "os"
    "time"
    "path"
    "strings"
    "strconv"
    "runtime"
    "net/http"
    "net/url"
    "io/ioutil"
    "encoding/json"
)

var config *Config

type DomainResponse struct {
    Status  Status
    Info    Info
    Domains []Domain
}

type RecordResponse struct {
    Status  Status
    Domain  Domain
    Records []Record
}

type Status struct {
    Code        string
    Message     string
    Created_at  string
}

type Info struct {
    Domain_total int
    All_total    int
    Mine_total   int
    Share_total  int
    Vip_total    int
    Ismark_total int
    Pause_total  int
    Error_total  int
    Lock_total   int
    Spam_total   int
    VipExpire    int
    ShareOut_total int
}
type Domain struct {
    Id   int
    Name string
}

type Record struct {
    Id      string
    Name    string
    Value   string
}

type Config struct {
    TargetDomains []string
    DnspodEmail  string
    DnspodPasswd string
}

func main() {
    config = parseConfigFile()
    fmt.Println("ddns started.")

    for {
        done := make(chan int, 1)
        go checkOrUpdate(done)

        ticker := time.NewTicker(5 * time.Second)
        select {
        case flag := <-done:
            if flag > 0 {
                fmt.Println("sleep 1 hour...")
                time.Sleep(1 * time.Hour)
            } else {
                time.Sleep(10 * time.Second)
            }
        case <-ticker.C:
        }
    }
}

func parseConfigFile() *Config {
    _, execDir, _, ok := runtime.Caller(0)
    if !ok {
        fmt.Println(execDir)
    }

    file := fmt.Sprintf("%s/../config.json", path.Dir(execDir))
    content, err := ioutil.ReadFile(file)

    if err != nil {
        fmt.Println("cannot read config file,", file, err)
        os.Exit(1)
    }

    var config Config
    json.Unmarshal(content, &config)
    return &config
}

func checkOrUpdate(done chan int) {
    nowIp := getNowIp()
    var deepSleep bool

    for _, domainName := range config.TargetDomains {
        domainId := getDomainId(domainName)
        recordId, recordIp, ds := getRecordIdAndRecordIp(domainName, domainId)
        if ds {
            deepSleep = ds
        }
        if len(recordId) < 1 {
            fmt.Println("domain ", domainName ," record not found, skip update")
            continue
        }

        if nowIp != recordIp {
            if err := updateTargetDomainDNS(domainName, domainId, recordId); err == nil {
                fmt.Printf("Success to update domain_name: %s, now ip: %s, record ip: %s\n", domainName, nowIp, recordIp)
            } else {
                fmt.Printf("Failed to update domain_name: %s, now ip: %s, record ip: %s\n", domainName, nowIp, recordIp)
            }
        } else {
            fmt.Printf("Domain %s no need to update, now ip: %s, record ip: %s\n", domainName, nowIp, recordIp)
        }
    }

    if deepSleep {
        done <- 1
        return
    }
    close(done)
}

func updateTargetDomainDNS(domainName string, domainId int, recordId string) (err error) {
    updateUrl := "https://dnsapi.cn/Record.Ddns"
    parms     := make(url.Values, 0)

    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain_id", strconv.Itoa(domainId))
    parms.Add("record_id", recordId)
    parms.Add("sub_domain", strings.Split(domainName, ".")[0])
    parms.Add("record_line", "默认")
    parms.Add("format", "json")

    var res *http.Response
    res, err = http.PostForm(updateUrl, parms)
    if err != nil {
        fmt.Println("cannot update damian dns", updateUrl)
        return
    }
    defer res.Body.Close()

    return err
}

func getDomainId(domainName string) (domainId int) {
    domainUrl := "https://dnsapi.cn/Domain.List"
    parms     := make(url.Values, 0)
    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("format", "json")

    res, err := http.PostForm(domainUrl, parms)
    if err != nil {
        fmt.Println("cannot get damian id via ", domainUrl)
        return
    }

    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("cannot get damian id via ", domainUrl)
        return
    }

    var dr DomainResponse
    json.Unmarshal(body, &dr)

    for _, domain := range dr.Domains {
        if strings.Index(domainName, domain.Name) != -1 {
            return domain.Id
        }
    }
    return
}

func getRecordIdAndRecordIp(domainName string, domainId int) (recordId, recordIp string, deepSleep bool) {
    recordUrl := "https://dnsapi.cn/Record.List"
    parms     := make(url.Values, 0)

    parms.Add("login_email", config.DnspodEmail)
    parms.Add("login_password", config.DnspodPasswd)
    parms.Add("domain_id", strconv.Itoa(domainId))
    parms.Add("format", "json")

    res, err := http.PostForm(recordUrl, parms)
    if err != nil {
        fmt.Println("cannot get record id via ", recordUrl)
        return
    }

    defer res.Body.Close()
    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("cannot get record id via ", recordUrl)
        return
    }

    var rr RecordResponse
    json.Unmarshal(body, &rr)
    if rr.Status.Code != "1" {
        fmt.Println(rr.Status.Message)
        return "", "",  true
    }

    for _, record := range rr.Records {
        if strings.Index(domainName, record.Name) != -1 {
            return record.Id, record.Value, false
        }
    }
    return
}

func getNowIp() string {
    ipUrl := "http://agideo.com/ip"
    res, err := http.Get(ipUrl)
    if err != nil {
        fmt.Println("cannot get current ip via ", ipUrl, err.Error())
        return ""
    }
    defer res.Body.Close()
    nowIp, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("read err, cannot get current ip via ", ipUrl, err.Error())
        return ""
    }

    return string(nowIp)
}
