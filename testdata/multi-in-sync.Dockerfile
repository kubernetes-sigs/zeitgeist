# This fake Dockerfile embeds versions so we can test pattern-matching
FROM debian

ENV TERRAFORM_VERSION=0.12.3

RUN docker pull gcr.io/kubernetes-helm/tiller:2.12.2

RUN echo TERRAFORM_VERSION=0.12.3
