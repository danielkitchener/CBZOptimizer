FROM alpine
LABEL authors="Belphemur"
ARG APP_PATH=/usr/local/bin/CBZOptimizer
ARG TARGETPLATFORM
ENV USER=abc
ENV CONFIG_FOLDER=/config
ENV PUID=99

RUN mkdir -p "${CONFIG_FOLDER}" && \
    adduser \
    -S \
    -H \
    -h "${CONFIG_FOLDER}" \
    -G "users" \
    -u "${PUID}" \
    "${USER}" && \
    chown ${PUID}:users "${CONFIG_FOLDER}"

COPY $TARGETPLATFORM/CBZOptimizer ${APP_PATH}

RUN apk add --no-cache \
    inotify-tools \
    bash \
    bash-completion && \
    chmod +x ${APP_PATH} && \
    ${APP_PATH} completion bash > /etc/bash_completion.d/CBZOptimizer

VOLUME ${CONFIG_FOLDER}
USER ${USER}
ENTRYPOINT ["/usr/local/bin/CBZOptimizer"]