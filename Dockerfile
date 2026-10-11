FROM golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/backend ./cmd/server

FROM scratch
COPY --from=builder /out/backend /backend
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/backend"]
