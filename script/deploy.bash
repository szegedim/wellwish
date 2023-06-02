#!/bin/bash

# Run this script on each node
# Create a load balancer TLS closure from SiteUrl:443 to metadata.Http11Port on each node.
# Nodes will take care of propagating the index for stateful bags

#docker pull php@sha256:b0eca9a9cb893d096dc0fc4a80a44697cabe6e1ed965cbccf5fd6046b4eda341
#docker pull node@sha256:14f0471d0478fbb9177d0f9e8c146dc872273dcdcfc7fea93a27ed81fc6b0e96
#docker pull mcr.microsoft.com/azure-functions/dotnet@sha256:9db3f0b48212872b5b52276a79e2175058d0340cc8412c57c482398312f99596


docker run -d --rm --restart=always --net=host -p 7777:7777 --name=stateful schmiedent/wellwish go run main.go
docker run -d --rm --restart=always --net=host --name=serverless0 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless1 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless2 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless3 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless4 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless5 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless6 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless7 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless8 schmiedent/wellwish go run burst/box/main.go
docker run -d --rm --restart=always --net=host --name=serverless9 schmiedent/wellwish go run burst/box/main.go
