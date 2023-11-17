FROM registry.datacommand.co.kr/golang:1.14 as builder

WORKDIR /build
COPY . .

ARG VERSION

RUN make clean build

FROM alpine

RUN apk --no-cache add tzdata

COPY --from=builder /build/cluster-manager /

ENTRYPOINT ["/cluster-manager"]
