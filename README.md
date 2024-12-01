# Webs.sh

## What is Webs.sh?
Webs.sh is a simple、fast reverse proxy to help you expose a local server behind a NAT or firewall to the Internet.

Using the ssh protocol means that Linux, macOS can quickly connect using the `ssh` command, and windows can connect using ?

```text
                                            Request                                                                                      
                            ┌────────────dispatch: foo────────────┐                                                                       
                            │                                     │                                                                       
                            │                                     │                                                                       
┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┼ ─ ─ ─ ─                    ┌ ─ ─ ─ ─│─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                                            
              Host Node     │        │                            │    Exit Node              │                                           
│                           ▼                            │        │                                                                       
   ┌───────────┐       ┌────────┐    │      .─────.          ┌────────┐        ┌──────────────┴───────┐        .─────.        ┌─────────┐
│  │   Node    │       │        │          ╱       ╲     │   │        │        │                      │       ╱       ╲       │         │
   │http://127.│◀ ─ ─ ▶│ client │◀═══╬═══▶(   SSH   )◀══════▶│ Server │◀ ─ ─ ─▶│http(s)://foo.webs.sh │◀ ─ ─▶(internet )◀─ ─ ▶│ Browser │
│  │0.0.1:3000 │       │        │          `.     ,'     │   │        │        │                      │       `.     ,'       │         │
   └───────────┘       └────────┘    │       `───'           └────────┘        └──────────────┬───────┘         `───'         └─────────┘
│                           │                            │        ▲                                                                       
                            │        │                            │                           │                                           
│                           │                            │        │                                                                       
 ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ─ ─ ─ ┘                    ─ ─ ─ ─ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘                                           
                            │             Response                │                                                                       
                            └────────────────OK───────────────────┘                                                                       
                                                                                                                                                   
```

## Feature
- Use ssh client, no need to install other clients
- Support macOS、Linux
- Windows putty

## Quick start
1. map the local listening port 8082 to the public network via webs.sh

```shell
$ ssh -R 443::8082 my.webs.sh
```

2. After establishing a successful remote SSH connection, the public access URL is returned `https://0u6rq7.webs.sh`

3. Access from the public network
```shell
$ curl https://0u6rq7.webs.sh
```

## Deploy
Preparation:
- domain name `webs.sh`
- server IP `157.xxx.xx.xx`

```shell
{
  "addr": ":443",
  "ssh_addr": ":22",
  "domain": "Webs.sh", # change you domain
  "idle_timeout": 300,
  "key": "-----BEGIN OPENSSH PRIVATE KEY-----\nxxxxxx\n-----END OPENSSH PRIVATE KEY-----\n"
}
```

1. Configure dns records
```shell
    A webs.sh 157.xxx.xx.xx
    A *.webs.sh 157.xxx.xx.xx
```
2. Get cloudflare [token](https://dash.cloudflare.com/profile/api-tokens) to the cf_token field of config.json

3. generator ssh key:
```shell
$ ssh-keygen -b 2048 -f Webs.sh_rsa
$ cat webs.sh_rsa #Copy the contents of the generated private key to the key field of config.json 
```
4. start Webs.sh

```shell
#.
#├── Webs.sh
#└── config.json
#

$  ./Webs.sh
```

## TODO
- Support windows

## Developer


## License
[BSD](LICENSE)

## Other
[Deepl](https://www.deepl.com/translator)