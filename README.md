# Slowpoke
[![Build Status](https://travis-ci.org/jamesbarnett91/slowpoke.svg?branch=master)](https://travis-ci.org/jamesbarnett91/slowpoke)

Slowpoke is a simple TCP proxy which can introduce configurable latency between packet delivery.
This allows you to test and profile how your application behaves with different levels of latency between services such as databases or caches.

## Running
If you have [Golang](https://golang.org) set up, you can install slowpoke using the standard `go get`package manager
```
go get github.com/jamesbarnett91/slowpoke
```
Otherwise, you can download one of the prebuilt binaries on the [releases page](https://github.com/jamesbarnett91/slowpoke/releases). 

Then invoke slowpoke passing the port on the local machine slowpoke should listen for connections on (`-p`), the host:port of the target (`-t`) and the duration of latency to apply between packets (`-l`)
E.g.
```
slowpoke -p 8181 -t localhost:8000 -l 25ms
```
will proxy traffic between `localhost:8181` and `localhost:8000` with 25 milliseconds of added latency between packets.

Possible options are :
```
Application Options:
  -p, --port=    The port Slowpoke should listen for connections on
  -t, --target=  The target address in host:port form
  -l, --latency= The duration of latency to apply to data packets, specified as a number and unit. E.g. 15ms or 2s. Supported units are 'us', 'ms', 's', 'm'
                 and 'h' (default: 0ms)
  -b, --buffer=  The size of the transfer buffer in bytes. Latency is applied between each buffer flush. Therefore total latency applied is equal to
                 '(totalDataTransferred/bufferSize) * latency' (default: 1500)
  -v, --verbose  Additional log verbosity. -v or -vv

Help Options:
  -h, --help     Show this help message
```

## Use cases
The primary use case for this tool is to simulate running you application on a slower network than the one you develop on. If you know your production environment has non-trivial latency on TCP connections, it's useful to run with these latencies on your development environment to ensure your application is still performant.

For example, in your production environment database connections may have to go through a VPN to cross different regions/AZs, which adds 5ms of latency to every packet.

By using Slowpoke you can simulate this 5ms overhead on your development env which would otherwise probably have near zero latency, especially if you run you DB on the same machine as your app connected via the loopback interface.
This may highlight areas of your application which perform poorly with this limitation, which otherwise would not have been apparent until deployed in production. 

E.g. you may have a screen which needs to deserialise 200 DB entities and currently fetches each entity in it's own query, rather than fetching them all in bulk. On your development env without the network latency this may have no noticeable performance impact, but with 5ms of packet latency that's now ~2 seconds added to your page load time and highlights something that needs to be fixed.


## Caveats
In order to keep Slowpoke simple and portable it doesn't actually add latency between each _packet_. Instead, it adds latency between writes of a configurably sized byte buffer. By default this buffer size is 1500 to match the standard ethernet MTU size, but can be changed depending on your use case.

Incidentally this also lets you simulate the MTU size change between running on the loopback interface (localhost) usually at the maximum 65536, and going over ethernet at 1500, which is sometimes another subtle difference between dev and prod deployments of applications.

Additionaly, because Slowpoke does not read TCP packets directly, it can't simulate latency on non 'data' packets (such as SYN/ACK).
