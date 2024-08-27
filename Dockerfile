FROM alpine:latest
LABEL authors="Belphemur"
ENV USER=abc
ENV CONFIG_FOLDER=/config
ENV PUID=99
ENV PGID=100
RUN mkdir -p "${CONFIG_FOLDER}" && addgroup -g "${PGID}" "${USER}" && adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "${USER}" \
    --no-create-home \
    --uid "${PUID}" \
    "${USER}" && \
    chown ${PUID}:${GUID} /config "${CONFIG_FOLDER}"

COPY CBZOptimizer /usr/local/bin/CBZOptimizer

RUN apk add --no-cache inotify-tools && chmod +x /usr/local/bin/CBZOptimizer

USER ${USER}
ENTRYPOINT ["/usr/local/bin/CBZOptimizer"]