#---------------------------------------------------------------------------------------------------
#
#---------------------------------------------------------------------------------------------------

ARG PLATFORM=linux/amd64
ARG GO_ARCH=amd64
ARG GO_ARM=7

FROM nginx:1.21.6-alpine

COPY ./cmd/nginx/config/nginx.conf /etc/nginx/conf.d/default.conf
