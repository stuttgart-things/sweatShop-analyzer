FROM golang:1.23 as builder
LABEL maintainer="Patrick Hermann patrick.hermann@sva.de"

ARG GO_MODULE="github.com/stuttgart-things/sweatShop-analyzer"
ARG VERSION=""
ARG BUILD_DATE=""
ARG COMMIT=""

WORKDIR /src/
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/sweatShop-analyzer \
    -ldflags="-X ${GO_MODULE}/internal.version=${VERSION} -X ${GO_MODULE}/internal.date=${BUILD_DATE} -X ${GO_MODULE}/internal.commit=${COMMIT}"

FROM eu.gcr.io/stuttgart-things/sthings-alpine@sha256:8d67f8b99f4bd4329cbdf5be80f8c8683a8a5cbe341c0860412c984b0a20a621
COPY --from=builder /bin/sweatShop-analyzer /bin/sweatShop-analyzer

ENTRYPOINT ["sweatShop-analyzer"]
