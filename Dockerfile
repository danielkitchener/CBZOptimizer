FROM alpine:latest
LABEL authors="Belphemur"
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

COPY CBZOptimizer /usr/local/bin/CBZOptimizer

RUN apk add --no-cache inotify-tools bash-completion libwebp && chmod +x /usr/local/bin/CBZOptimizer && /usr/local/bin/CBZOptimizer completion bash > /etc/bash_completion.d/CBZOptimizer

VOLUME ${CONFIG_FOLDER}
USER ${USER}
ENTRYPOINT ["/usr/local/bin/CBZOptimizer"]