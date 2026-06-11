# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags='-s -w' -o /out/aggregator ./cmd/aggregator

FROM scratch

COPY --from=build /out/aggregator /aggregator

WORKDIR /work
USER 65532:65532

ENTRYPOINT ["/aggregator"]
CMD ["--help"]
