# coredns

## desc
在reserver&mask测试过程中，我们需要用到dns工具，这里使用coredns

## install
* 获取docker镜像
```bash
docker pull coredns/coredns
```

## example
具体使用参考: https://github.com/coredns/coredns, 这里我们提供robot reserve测试用例。
在172.168.3.59这台机器上运行dns服务器，提供一个次级域名ontsnip.com，
以及两个子域名reserve1.ontsnip.com, reserve2.ontsnip.com.

#### 编写Corefile
```dtd
. {
    log
    errors
    auto
    reload 10s
    forward . /etc/resolv.conf 
    file /root/db.ontsnip.com ontsnip.com
}
```

#### 编写SOA
```dtd
$ORIGIN ontsnip.com.
@	3600 IN	SOA sns.dns.icann.org. dylenfu@126.com. (
				2020052604 ; serial
				60         ; refresh in seconds (1 min)
				3600       ; retry (1 hour)
				1209600    ; expire (2 weeks)
				3600       ; minimum (1 hour)
				)

	3600 IN NS a.iana-servers.net.
	3600 IN NS b.iana-servers.net.

reserve1    IN A     172.168.3.158
reserve2    IN A     172.168.3.163
```

#### docker运行

```bash
#!/bin/bash

workspace=/home/ubuntu/docker/coredns

docker stop coredns
docker rm coredns
docker run -d --name coredns \
-v=$workspace:/root/ \
-p 53:53/udp \
coredns/coredns -conf /root/Corefile

docker logs -f coredns
```

#### 添加dns服务器

在172.168.3.158和172.168.3.163这些需要用到该dns服务的机器上修改/etc/resolv.conf,添加
```dtd

nameserver 172.168.3.59
```

#### 测试

在172.168.3.158这台机器上使用dig命令分析，要能看到ANSWER SECTION下具体域名及地址.
```bash
root@ubuntu-virtual-machine:~# dig reserve1.ontsnip.com

; <<>> DiG 9.11.3-1ubuntu1.12-Ubuntu <<>> reserve1.ontsnip.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 55664
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 2, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
; COOKIE: b5e100b061abf199 (echoed)
;; QUESTION SECTION:
;reserve1.ontsnip.com.		IN	A

;; ANSWER SECTION:
reserve1.ontsnip.com.	3600	IN	A	172.168.3.158

;; AUTHORITY SECTION:
ontsnip.com.		3600	IN	NS	a.iana-servers.net.
ontsnip.com.		3600	IN	NS	b.iana-servers.net.

;; Query time: 0 msec
;; SERVER: 172.168.3.59#53(172.168.3.59)
;; WHEN: Tue Jun 02 18:03:26 CST 2020
;; MSG SIZE  rcvd: 183
```
