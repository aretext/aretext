ARG BASE_IMAGE=base-image
ARG BUILD_VERSION=1
FROM --platform=$TARGET_PLATFORM ${BASE_IMAGE} AS build_stage
ENV APP_HOME=/workspace APP_VERSION=${BUILD_VERSION}
LABEL description="generic image"
RUN echo "$APP_HOME" && echo ${APP_VERSION}
