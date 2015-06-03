## DDNS

[![Build Status](https://github.com/c77cc/ddns)](https://github.com/c77cc/ddns)

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
