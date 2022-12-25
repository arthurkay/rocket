# How to run your own rocketd server

Running your own rocket server is really easy! The instructions below will guide you along your way!

## 1. Get an SSL certificate

rocket provides secure tunnels via TLS, so you'll need an SSL certificate. Assuming you want to create
tunnels on *.livingopensource.africa, buy a wildcard SSL certificate for *.livingopensource.africa. Note that if you
don't need to run https tunnels that you don't need a wildcard certificate. (In fact, you can
just use a self-signed cert at that point, see the section on that later in the document).

## 2. Modify your DNS

You need to use the DNS management tools given to you by your provider to create an A
record which points *.livingopensource.africa to the IP address of the server where you will run rocketd.

## 3. Compile it

You can compile an rocketd server with the following command:

    make server

Or you can download it from release section https://github.com/arthurkay/rocket/releases you need **rocketd** file

Then copy the binary over to your server.

## 4. Run the server


### Some important options

#### Specifying your TLS certificate and key

rocket only makes TLS-encrypted connections. When you run rocketd, you'll need to instruct it
where to find your TLS certificate and private key. Specify the paths with the following switches:

    -tlsKey="/path/to/tls.key" -tlsCrt="/path/to/tls.crt"

#### Setting the server's domain

When you run your own rocketd server, you need to tell rocketd the domain it's running on so that it
knows what URLs to issue to clients.

    -domain="livingopensource.africa"

#### Protect you client(rocket) to server(rocketd) connection with a CA

if you include this, all rocket client will need two aditional arguments to connect to this server, that needs to be a client certificate for this CA

     -tunnelTLSClientCA=./ca.crt

to see which parameters rocket client needs, is in "Connect with a client" section

#### Protect you exposed subdomain with a CA

ej: sample.livingopensource.africa will only be accesible when the user has a client certificate allowed by this CA and also needs to be installed on his device

     -tlsClientCA=./ca.crt

### (Option 1) Command line
You'll run the server with the following command.

    ./rocketd -domain livingopensource.africa -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=./tls.crt -tlsKey=./tls.key


### (Option 2) Supervidord
you can use supervisor to run in background, here is a sample config file

```ini
[program:rocketd]
directory=/root/rocket
autostart=true
autorestart=true
command=/root/rocket/rocketd -domain livingopensource.africa -log-level=WARNING -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=./certs/tls.crt -tlsKey=./certs/tls.key

```

### (Option 3) Docker Compose

or also you can use docker-compose

```yaml
version: "3.7"

services:
  rocketd:
    image: nerdygeek/rocketd
    entrypoint: rocketd
    command: -domain livingopensource.africa -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=/certs/tls.crt -tlsKey=/certs/tls.key
    volumes:
      - /home/certs:/certs
    ports:
      - 80:80
      - 443:443
      - 4443:4443
```



## 5. Configure client

In order to Configure client, you'll need to set two options in rocket's configuration file.
The rocket configuration file is a simple YAML file that is read from ~/.rocket by default. You may specify
a custom configuration file path with the -config switch. Your config file must contain the following two
options.

    server_addr: livingopensource.africa:4443
    trust_host_root_certs: true

Substitute the address of your rocketd server for "livingopensource.africa:4443". The "trust_host_root_certs" parameter instructs
rocket to trust the root certificates on your computer when establishing TLS connections to the server. By default, rocket
only trusts the root certificate for livingopensource.africa.

## 6. Connect with a client

Then, just run rocket as usual to connect securely to your own rocketd server!

    rocket -subdomain=customsubdomain 127.0.0.1:3000

or you can specify a custom server here too

    rocket -log=stdout -serveraddr=livingopensource.africa:4443 -subdomain=customsubdomain 127.0.0.1:3000

or if you are using "Protect you client(rocket) to server(rocketd) connection with a CA
" here is where you need to include to arguments ` -tlsClientCrt=./client_crt -tlsClientKey=./client_key`

    rocket -log=stdout -log-level=WARNING -serveraddr=ejemplo:4443 -tlsClientCrt=./client_crt -tlsClientKey=./client_key -subdomain=customsubdomain 127.0.0.1:3000
# rocketd with a self-signed SSL certificate

It's possible to run rocketd with a a self-signed certificate, but you'll need to recompile rocket with your signing CA.
If you do choose to use a self-signed cert, please note that you must either remove the configuration value for
trust_host_root_certs or set it to false:

    trust_host_root_certs: false
