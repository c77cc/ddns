## DDNS

[![Build Status](https://drone.io/github.com/c77cc/ddns/status.png)](https://drone.io/github.com/c77cc/ddns/latest)

动态更新域名解析记录，目前只支持DNSPOD

## Build

* Get source code from Github:

```bash
git clone https://github.com/c77cc/ddns
```

```bash
cd ddns
go build

修改账户密码和域名
cp config.json.default config.json

./ddns
```

## 独立部署
```bash
cp ddns.initd.sh /etc/init.d/ddns && chkconfig --add ddns && chkconfig ddns on
```

## supervisor部署
```bash
yum install supervisor -y
cat ddns.supervisord.confg >> /etc/supervisord.conf && /etc/init.d/supervisord restart
```
