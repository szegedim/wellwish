#!/bin/bash

# Repeat this on each node by copying it to the end of any startup script
# Create a load balancer TLS closure from SiteUrl:443 to 7777 on each node.
# Nodes will take care of propagating the index for stateful sacks

docker run -d --rm --restart=always --net=host -v /tmp/containers:/tmp/containers:rw -p 7777:7777 --name=stateful schmiedent/wellwish go run main.go

# docker run -d --rm --restart=always --net=none --privileged ... read apikeys from /tmp/containers/container*.metal and restart the ones where the token expired

docker run -d --rm --restart=always --net=host -v /tmp/containers/container0.metal:/tmp/containers/container0.metal:ro --name=container0 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container1.metal:/tmp/containers/container1.metal:ro --name=container1 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container2.metal:/tmp/containers/container2.metal:ro --name=container2 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container3.metal:/tmp/containers/container3.metal:ro --name=container3 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container4.metal:/tmp/containers/container4.metal:ro --name=container4 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container5.metal:/tmp/containers/container5.metal:ro --name=container5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container5.metal:/tmp/containers/container5.metal:ro --name=container5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container6.metal:/tmp/containers/container6.metal:ro --name=container6 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container7.metal:/tmp/containers/container7.metal:ro --name=container7 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container8.metal:/tmp/containers/container8.metal:ro --name=container8 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/container9.metal:/tmp/containers/container9.metal:ro --name=container9 schmiedent/wellwish go run burst/box/main.go
