FROM alpine:3.9 AS build

ENV HELM_VERSION="v2.10.0"
ENV TERRAFORM_VERSION="0.10.0"

RUN apk add --update ca-certificates git bash curl docker jq \
  && curl -LO https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
  && curl -L https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar xz && helm init --client-only
