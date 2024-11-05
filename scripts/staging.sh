#!/bin/bash

export SLUG=ghcr.io/awakari/source-sse
export VERSION=latest
docker tag awakari/source-sse "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
