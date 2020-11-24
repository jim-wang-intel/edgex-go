#  ----------------------------------------------------------------------------------
#  Copyright 2020 Intel Corp.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
#
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------

ARG BUILDER_BASE=golang:1.15-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-bootstrapper/security-bootstrapper

FROM alpine

RUN apk update && apk add dumb-init openssl \
    && rm -rf /var/cache/apk/* 

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2020 Intel Corporation'


ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

WORKDIR /edgex

RUN mkdir -p /edgex-init && \
    mkdir /edgex-staging

COPY --from=builder /edgex-go/cmd/security-bootstrapper/vault_wait_install.sh /edgex-staging
RUN chmod +x /edgex-staging/vault_wait_install.sh

COPY --from=builder /edgex-go/cmd/security-bootstrapper/security-bootstrapper .
COPY --from=builder /edgex-go/cmd/security-bootstrapper/res/configuration.toml ./res/

COPY --from=builder /edgex-go/cmd/security-bootstrapper/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh \
    && ln -s /usr/local/bin/entrypoint.sh /

ENTRYPOINT ["entrypoint.sh"]
CMD ["--init"]
