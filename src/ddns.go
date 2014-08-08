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

var (
    targetDomain string
    dnspodEmail  string
    dnspodPasswd string
)

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
    TargetDomain string
    DnspodEmail  string
    DnspodPasswd string
}

func main() {
    parseConfigFile()
    ticker := time.NewTicker(30 * time.Second)
    for _ = range ticker.C {
        checkOrUpdate()
    }
}

func parseConfigFile() {
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

    targetDomain = config.TargetDomain
    dnspodEmail  = config.DnspodEmail
    dnspodPasswd = config.DnspodPasswd
}

func checkOrUpdate() {
    nowIp    := getNowIp()
    domainId := getDomainId()
    recordId, recordIp := getRecordIdAndRecordIp(domainId)

    if nowIp != recordIp {
        fmt.Println("Starting update domain dns...")
        fmt.Printf("now ip: %s, record ip: %s, target_domain: %s, domain_id: %d, record_id: %s\n", nowIp, recordIp, targetDomain, domainId, recordId)
        updateTargetDomainDNS(domainId, recordId)
        fmt.Println("Success to update")
    } else {
        fmt.Printf("now ip: %s, record ip: %s\n", nowIp, recordIp)
        fmt.Println("No need update")
    }
}

func updateTargetDomainDNS(domainId int, recordId string) error {
    updateUrl := "https://dnsapi.cn/Record.Ddns"
    parms     := make(url.Values, 0)

    parms.Add("login_email", dnspodEmail)
    parms.Add("login_password", dnspodPasswd)
    parms.Add("domain_id", strconv.Itoa(domainId))
    parms.Add("record_id", recordId)
    parms.Add("sub_domain", strings.Split(targetDomain, ".")[0])
    parms.Add("record_line", "默认")
    parms.Add("formst", "json")

    _, err := http.PostForm(updateUrl, parms)
    if err != nil {
        fmt.Println("cannot update damian dns", updateUrl)
        return err
    }

    return nil
}

func getDomainId() (domainId int) {
    domainUrl := "https://dnsapi.cn/Domain.List"
    parms     := make(url.Values, 0)
    parms.Add("login_email", dnspodEmail)
    parms.Add("login_password", dnspodPasswd)
    parms.Add("format", "json")

    res, err := http.PostForm(domainUrl, parms)
    if err != nil {
        fmt.Println("cannot get damian id via ", domainUrl)
        return
    }

    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("cannot get damian id via ", domainUrl)
        return
    }

    var dr DomainResponse
    json.Unmarshal(body, &dr)

    for _, domain := range dr.Domains {
        if strings.Index(targetDomain, domain.Name) != -1 {
            return domain.Id
        }
    }
    return
}

func getRecordIdAndRecordIp(domainId int) (recordId, recordIp string) {
    recordUrl := "https://dnsapi.cn/Record.List"
    parms     := make(url.Values, 0)

    parms.Add("login_email", dnspodEmail)
    parms.Add("login_password", dnspodPasswd)
    parms.Add("domain_id", strconv.Itoa(domainId))
    parms.Add("format", "json")

    res, err := http.PostForm(recordUrl, parms)
    if err != nil {
        fmt.Println("cannot get record id via ", recordUrl)
        return
    }

    body, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("cannot get record id via ", recordUrl)
        return
    }

    var rr RecordResponse
    json.Unmarshal(body, &rr)

    for _, record := range rr.Records {
        if strings.Index(targetDomain, record.Name) != -1 {
            return record.Id, record.Value
        }
    }
    return
}

func getNowIp() string {
    ipUrl := "http://agideo.com/ip"
    res, err := http.Get(ipUrl)
    if err != nil {
        fmt.Println("cannot get current ip via ", ipUrl)
        return ""
    }
    nowIp, err1 := ioutil.ReadAll(res.Body)
    if err1 != nil {
        fmt.Println("cannot get current ip via ", ipUrl)
        return ""
    }

    return string(nowIp)
}
