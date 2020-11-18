FROM golang:1.15-buster  as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOBIN=/bin
WORKDIR /build
COPY go.mod go.sum ./
COPY internal ./internal
RUN go build ./internal/pkg/imports
COPY . .
RUN go build -o /bin/fake-cross-nse .


FROM alpine as runtime
COPY --from=build /bin/fake-cross-nse /bin/fake-nsmgr
CMD /bin/fake-nsmgr