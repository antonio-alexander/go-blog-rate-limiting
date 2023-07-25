#---------------------------------------------------------------------------------------------------
# go-blog-rate-limiting [Dockerfile]
# 
# Reference: https://stackoverflow.com/questions/63178036/how-to-find-commit-hash-from-within-a-running-docker-image
# commit: git rev-parse HEAD
# 
# https://stackoverflow.com/questions/6245570/how-to-get-the-current-branch-name-in-git
# branch: git rev-parse --abbrev-ref HEAD
# 
# https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
# The GOOS will always be linux, but in case you were curious
# 
# Sample docker build commands:
#  docker build -f ./cmd/Dockerfile . -t ghcr.io/antonio-alexander/go-blog-rate-limiting:amd64_latest --build-arg GIT_COMMIT=$GITHUB_SHA --build-arg GIT_BRANCH=$GITHUB_REF --build-arg PLATFORM=linux/amd64 --build-arg GO_ARCH=amd64
#  docker build -f ./cmd/Dockerfile . -t ghcr.io/antonio-alexander/go-blog-rate-limiting:armv7_latest --build-arg GIT_COMMIT=$GITHUB_SHA --build-arg GIT_BRANCH=$GITHUB_REF --build-arg PLATFORM=linux/arm/v7 --build-arg GO_ARCH=arm --build-arg GO_ARM=7
#
#---------------------------------------------------------------------------------------------------

ARG PLATFORM=linux/amd64
ARG GO_ARCH=amd64
ARG GO_ARM=7

FROM golang:alpine AS build

ARG GO_ARCH
ARG GO_ARM

WORKDIR /go/src/go-blog-rate-limiting

COPY . /go/src/go-blog-rate-limiting

RUN apk add jq

RUN \
    VERSION=`cat /go/src/go-blog-rate-limiting/version.json | jq '.Version' | sed 's/"//g'` \
    && cd cmd/client \
    && env GOARCH=${GO_ARCH} GOARM=${GO_ARM} GOOS=linux go build -o go-blog-rate-limiting \
    && chmod +X go-blog-rate-limiting \
    && tar -czvf go-blog-rate-limiting.tar.gz ./go-blog-rate-limiting

FROM --platform=${PLATFORM} alpine:latest

ARG GIT_BRANCH
ARG GIT_COMMIT

WORKDIR /app

COPY --from=build /go/src/go-blog-rate-limiting/cmd/client/go-blog-rate-limiting.tar.gz /app/go-blog-rate-limiting.tar.gz

LABEL org.opencontainers.image.source=https://github.com/antonio-alexander/go-blog-rate-limiting

CMD tar -xzf go-blog-rate-limiting.tar.gz && ./go-blog-rate-limiting