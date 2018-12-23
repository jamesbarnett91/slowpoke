# Slowpoke
Slowpoke is a simple TCP proxy which can introduce a configurable latency between packet transfers.
This allows you to test and profile how your application behaves with different levels of latency between services such as databases or caches.

# Running
TODO once binaries built

# Use cases
The primary use case for this tool is to simulate running you application on a slower network than the one you develop on. If you know your production environment has non-trivial latency on TCP connections, it's useful to run with these latencies on your development environment to ensure your application is still performant.

For example, in your production environment database connections may have to go through a VPN to cross different regions/AZs, which adds 5ms of latency to every packet.
By using Slowpoke you can simulate this 5ms overhead on your development env which would otherwise probably have near zero latency, especially if you run you DB on the same machine as your app connected via the loopback interface.
This may highlight areas of your application which perform poorly with this limitation, which otherwise would not have been apparent until deployed in production. 
E.g. you may have a screen which needs to deserialise 200 DB entities and currently fetches each entity in it's own query, rather than fetching them all in bulk. On your development env without the network latency this may have no noticeable performance impact, but with 5ms of packet latency that's now ~2 seconds added to your page load time and highlights something that needs to be fixed.


# Note on packets
In order to keep Slowpoke simple and portable it doesn't actually add latency between each _packet_. Instead, it adds latency between writes of a configurably sized byte buffer. By default this buffer size is 1500 to match the standard ethernet MTU size, but can be changed depending on your use case.

Incidentally this also lets you simulate the MTU size change between running on the loopback interface (localhost) usually at the maximum 65536, and going over ethernet at 1500, which is sometimes another subtle difference between dev and prod deployments of applications.

