# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.12-alpine as builder

RUN apk --update add git upx

WORKDIR /go/src/github.com/pluies/zeitgeist
ENV GO111MODULE=on
ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o zeitgeist .

# Make things even smaller!
RUN upx zeitgeist

FROM scratch
# Copy trusted CAs for TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# And our executable!
COPY --from=builder /go/src/github.com/pluies/zeitgeist/zeitgeist /
# Run as non-root
USER 1001
CMD ["/zeitgeist"]
