# 4Chain

## What is 4chain?
4chain is a simple、fast reverse proxy to help you expose a local server behind a NAT or firewall to the Internet. 

Using the ssh protocol means that Linux, macOS can quickly connect using the `ssh` command, and windows can connect using [PuTTY](https://www.putty.org/)

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

## Quick start
local server listen port: `8082`, remote port: `80`or `443`

```shell
ssh -R 443::8082 4chain.me
```

## Developer


## License
[BSD](LICENSE)

