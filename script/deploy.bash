#!/bin/bash

# Run this script on each node
# Create a load balancer TLS closure from SiteUrl:443 to 7777 on each node.
# Nodes will take care of propagating the index for stateful sacks

docker run -d --rm --restart=always --net=host -v /tmp/containers:/tmp/containers:rw -p 7777:7777 --name=stateful schmiedent/wellwish go run main.go

# docker run -d --rm --restart=always --net=none --privileged ... read apikeys from /tmp/containers/container*.metal and restart the ones where the token expired

docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless0.metal:/tmp/containers/stateless0.metal:ro --name=stateless0 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless1.metal:/tmp/containers/stateless1.metal:ro --name=stateless1 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless2.metal:/tmp/containers/stateless2.metal:ro --name=stateless2 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless3.metal:/tmp/containers/stateless3.metal:ro --name=stateless3 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless4.metal:/tmp/containers/stateless4.metal:ro --name=stateless4 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless5.metal:/tmp/containers/stateless5.metal:ro --name=stateless5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless5.metal:/tmp/containers/stateless5.metal:ro --name=stateless5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless6.metal:/tmp/containers/stateless6.metal:ro --name=stateless6 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless7.metal:/tmp/containers/stateless7.metal:ro --name=stateless7 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless8.metal:/tmp/containers/stateless8.metal:ro --name=stateless8 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host -v /tmp/containers/stateless9.metal:/tmp/containers/stateless9.metal:ro --name=stateless9 schmiedent/wellwish go run burst/box/main.go
