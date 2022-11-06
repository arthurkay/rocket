# rocket - Edge Network Computing

### ”I want to expose a local server behind a NAT or firewall to the internet.”

# Install client

Install supports **Linux** and **MacOS** 

```bash
rocket -subdomain=customsubdomain 3000
```
sample output

```bash
rocket                                                           (Ctrl+C to quit)
                                                                                
Tunnel Status                 online                                            
Version                       1.0/1.0                                           
Forwarding                    http://customsubdomain.livingopensource.africa -> 127.0.0.1:3000            
Forwarding                    https://customsubdomain.livingopensource.africa -> 127.0.0.1:3000           
Web Interface                 http://127.0.0.1:4040                             
# Conn                        0                                                 
Avg Conn Time                 0.00ms 
```

# Downloads

just download in [Release section](https://github.com/arthurkay/rocket/releases)

## What is rocket?

rocket is a reverse proxy that creates a secure tunnel from a public endpoint to a locally running web service.
rocket captures and analyzes all traffic over the tunnel for later inspection and replay.

## What can I do with rocket?

- Expose any http service behind a NAT or firewall to the internet on a subdomain of livingopensource.africa
- Expose any tcp service behind a NAT or firewall to the internet on a random port of livingopensource.africa
- Inspect all http requests/responses that are transmitted over the tunnel
- Replay any request that was transmitted over the tunnel

## What is rocket useful for?

- Temporarily sharing a website that is only running on your development machine
- Demoing an app at a hackathon without deploying
- Developing any services which consume webhooks (HTTP callbacks) by allowing you to replay those requests
- Debugging and understanding any web service by inspecting the HTTP traffic
- Running networked services on machines that are firewalled off from the internet

## Developing on rocket

[rocket developer's guide](docs/DEVELOPMENT.md)

## Disclaimer

rocket is a spinoff fork of ngrok
