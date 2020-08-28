FROM golang:1.13-buster  as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOBIN=/bin
WORKDIR /build
RUN go get github.com/go-delve/delve/cmd/dlv@v1.4.0
COPY go.mod go.sum ./
COPY pkg ./pkg
RUN go build ./pkg/imports
COPY . .
RUN go build -o /bin/proxy-endpoint .

EXPOSE 6000

FROM alpine as runtime
COPY --from=build /bin/dlv /bin/dlv
COPY --from=build /bin/proxy-endpoint /bin/proxy-endpoint
CMD /bin/proxy-endpoint