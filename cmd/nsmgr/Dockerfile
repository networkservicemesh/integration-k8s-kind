FROM golang:1.13-buster  as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOBIN=/bin
WORKDIR /build
COPY go.mod go.sum ./
COPY pkg ./pkg
RUN go build ./pkg/imports
COPY . .
RUN go build -o /bin/nsmgr .


FROM alpine as runtime
COPY --from=build /bin/nsmgr /bin/nsmgr
CMD /bin/nsmgr