Tic-tac-toe game
================

## TL;DR

        make && ./tic-tac-toe

and install `ssl/rootCA.crt` into your browser cert registry to avoid ssl warnings        

service will be available on https://localhost/api/v1/games

## Build

For build with defaults run 
        
        make

To set custom cert altNames run `ssl/gencerts.sh` with parameters

generate CA

        ssl/gencerts.sh ca  

generate service cert

        ssl/gencerts.sh service -a 192.168.0.101 -h service-dns-name


Install `ssl/rootCA.crt` into your browser cert registry to avoid ssl warnings

### gencerts.sh usage

```
usage: gencerts.sh ca | service -h <host> -a <ip_addr> | ca-show | service-show

Commands:
  ca - generate CA cert

  service - generate service cert
      host        alternative dns name for cert (default: localhost)
      ipaddr      alternative ip for cert (default: 127.0.0.1)

  ca-show - show CA cert

  service-show - show service cert

  clean - remove all certs, requests and other generated artefacts
```


## Run

To run `tic-tac-toe` with default parameters, run

        ./tic-tac-toe
        
or specify parameter from usage help below

```
Usage of ./tic-tac-toe:
  -addr string
    	TCP address to listen to (default "0.0.0.0:443")
  -cert string
    	path to tls-cert file (default "ssl/cert.pem")
  -debug
    	print debug messages
  -key string
    	path to tls-key file (default "ssl/key.pem")
  -storagePath string
    	path to storage with game files (default "storage")
```
