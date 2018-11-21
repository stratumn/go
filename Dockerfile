# Build a base image that contains the repository's code and its dependencies.
# This base image can be used to build go plugins for validation scripts.

ARG golang_image
FROM $golang_image

ARG org
ARG repository

RUN mkdir -p /go/src/github.com/$org/$repository
ADD . /go/src/github.com/$org/$repository