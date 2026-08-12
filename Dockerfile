FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/dotenvai ./cmd/dotenvai

FROM alpine:3.22
RUN apk add --no-cache git ca-certificates
COPY --from=build /out/dotenvai /usr/local/bin/dotenvai
COPY entrypoint.sh /entrypoint.sh
COPY THIRD_PARTY_NOTICES.md /licenses/THIRD_PARTY_NOTICES.md
ENTRYPOINT ["/entrypoint.sh"]
