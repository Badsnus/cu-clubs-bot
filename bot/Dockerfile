FROM golang:1.23.5-alpine AS builder

# Add build arguments for PR information
ARG PR_NAME="Unknown"
ARG PR_DESC="Unknown"
ARG BUILD_DATE="Unknown"

RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o bin/bot ./cmd

FROM alpine:3.19 AS runner

# Pass build arguments to the runner stage
ARG PR_NAME
ARG PR_DESC
ARG BUILD_DATE

# Set environment variables for version tracking
ENV BOT_PR_NAME=${PR_NAME}
ENV BOT_PR_DESC=${PR_DESC}
ENV BOT_BUILD_DATE=${BUILD_DATE}

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /opt

COPY --from=builder /opt/bin/bot ./

COPY telegram.yml /opt/telegram.yml
COPY locales /opt/locales
COPY logo.png /opt/logo.png
COPY mail.html /opt/mail.html

CMD ["./bot"]