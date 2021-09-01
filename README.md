# 4Chain

## What is 4chain?
4chain is a simple、fast reverse proxy to help you expose a local server behind a NAT or firewall to the Internet. 

Using the ssh protocol means that Linux, macOS can quickly connect using the `ssh` command, and windows can connect using ?

```text
                                               Request                                                                                      
                            ┌───────────────dispatch: foo───────────┐                                                                       
                            │                                       │                                                                       
                            │                                       │                                                                       
┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─                      ┌ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                                            
              Host Node     │        │                              │    Exit Node              │                                           
│                           ▼                              │        │                                                                       
   ┌───────────┐       ┌────────┐    │      .─────.            ┌────────┐        ┌──────────────┴────────┐        .─────.        ┌─────────┐
│  │   Node    │       │        │          ╱       ╲       │   │        │        │                       │       ╱       ╲       │         │
   │http://127.│◀ ─ ─ ▶│ client │◀═══╬═══▶(   SSH   )◀════════▶│ Server │◀ ─ ─ ─▶│http(s)://foo.4chain.me│◀ ─ ─▶(internet )◀─ ─ ▶│ Browser │
│  │0.0.1:3000 │       │        │          `.     ,'       │   │        │        │                       │       `.     ,'       │         │
   └───────────┘       └────────┘    │       `───'             └────────┘        └──────────────┬────────┘         `───'         └─────────┘
│                           │                              │        ▲                                                                       
     Private IPv4           │        │                              │                           │                                           
│                           │                              │        │                                                                       
 ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ┘                      ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘                                           
                            │             Response                  │                                                                       
                            └────────────────OK─────────────────────┘                                                                       
                                                                                                                                                   
```

## Feature
- Use ssh client, no need to install other clients
- Use [let's encrypt](https://letsencrypt.org/) via [certmagic](https://github.com/caddyserver/certmagic) tool
- Support macOS、Linux

## Quick start
1. map the local listening port 8082 to the public network via 4chain.me

```shell
$ ssh -R 443::8082 4chain.me
```
2. After establishing a successful remote SSH connection, the public access URL is returned `https://0u6rq7.4chain.me`

3. Access from the public network
```shell
    curl https://0u6rq7.4chain.me
```

## Deploy
Preparation: 
- domain name `4chain.me`
- server IP `157.xxx.xx.xx`

```shell
{
  "addr": ":443",
  "email": "admin@4chain.me",
  "cf_token": "xxx",
  "ssh_addr": ":22",
  "domain": "4chain.me", # change you domain
  "idle_timeout": 300,
  "key": "-----BEGIN OPENSSH PRIVATE KEY-----\nxxxxxx\n-----END OPENSSH PRIVATE KEY-----\n"
}
```

1. Configure cloudflare dns resolution
```shell
    A 4chain.me 157.xxx.xx.xx
    A *.4chain.me 157.xxx.xx.xx
```
2. Get cloudflare [token](https://dash.cloudflare.com/profile/api-tokens) to the cf_token field of config.json

3. generator ssh key:
```shell
$ ssh-keygen -b 2048 -f 4chain_rsa
$ cat 4chain_rsa #Copy the contents of the generated private key to the key field of config.json 
```
4. start 4chain

```shell
#.
#├── 4chain
#└── config.json
#

$  ./4chain
```

## TODO
- Support windows

## Developer


## License
[BSD](LICENSE)

## Other
[Deepl](https://www.deepl.com/translator)
