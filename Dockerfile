FROM debian:sid-slim
LABEL authors="Belphemur"
ARG APP_PATH=/usr/local/bin/CBZOptimizer
ENV USER=abc
ENV CONFIG_FOLDER=/config
ENV PUID=99

RUN mkdir -p "${CONFIG_FOLDER}" && \
    useradd \
        --system \
        --no-create-home \
        --home-dir "${CONFIG_FOLDER}" \
        --gid "users" \
        --uid "${PUID}" \
        "${USER}" && \
        chown ${PUID}:users "${CONFIG_FOLDER}"

COPY CBZOptimizer ${APP_PATH}

RUN apt-get update && \
    apt-get full-upgrade -y && \
    apt-get install -y inotify-tools bash-completion libwebp7 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    chmod +x ${APP_PATH} && \
    ${APP_PATH} completion bash > /etc/bash_completion.d/CBZOptimizer

VOLUME ${CONFIG_FOLDER}
USER ${USER}
ENTRYPOINT ["/usr/local/bin/CBZOptimizer"]