# syntax=docker/dockerfile:1

FROM golang:1.23.1 AS build-stage

WORKDIR /app

COPY /cmd/go.mod /cmd/go.sum ./

RUN go mod donwload

COPY /cmd/*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker

FROM build-stage AS rest-stege

RUN go test -v ./..

FROM grc.io/distroless/base-debian11 AS final-stage

WORKDIR /

COPY --from=build-stage /doc-image /doc-image

USER nonroot:nonroot

ENTRYPOINT ["/docker"]