FROM alpine:latest
LABEL authors="Belphemur"
ENV USER=abc
ENV CONFIG_FOLDER=/config
ENV PUID=99
RUN mkdir -p "${CONFIG_FOLDER}" && adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "users" \
    --no-create-home \
    --uid "${PUID}" \
    "${USER}" && \
    chown ${PUID}:${GUID} "${CONFIG_FOLDER}"

COPY CBZOptimizer /usr/local/bin/CBZOptimizer

RUN apk add --no-cache inotify-tools && chmod +x /usr/local/bin/CBZOptimizer && /usr/local/bin/CBZOptimizer completion bash > /etc/bash_completion.d/CBZOptimizer

USER ${USER}
ENTRYPOINT ["/usr/local/bin/CBZOptimizer"]