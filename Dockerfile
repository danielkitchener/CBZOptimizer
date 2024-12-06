FROM alpine:latest
LABEL authors="Belphemur"
ARG APP_PATH=/usr/local/bin/CBZOptimizer
ENV USER=abc
ENV CONFIG_FOLDER=/config
ENV PUID=99
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "users" \
    --uid "${PUID}" \
    --home "${CONFIG_FOLDER}" \
    "${USER}" && \
    chown ${PUID}:${GUID} "${CONFIG_FOLDER}"

COPY CBZOptimizer ${APP_PATH}

RUN apk add --no-cache inotify-tools bash-completion libwebp &&  \
    chmod +x ${APP_PATH} && \
    ${APP_PATH} completion bash > /etc/bash_completion.d/CBZOptimizer

VOLUME ${CONFIG_FOLDER}
USER ${USER}
ENTRYPOINT ["${APP_DATA}"]