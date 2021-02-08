#FROM --platform=${BUILDPLATFORM} golang:1.16-rc AS build
FROM golang:1.16-rc AS build
WORKDIR /src
#ENV CGO_ENABLED=0
COPY examples/hello .
ARG GOOS
ARG GOARCH
# for gitlab private repos, authentication with tokens is necessary 
# and GOPRIVATE must skip verification for git.cocus.com
RUN echo "machine git.cocus.com login go-build password 6H93rHDDn-HnL26tTtRP" > ~/.netrc
ENV GOPRIVATE=git.cocus.com/*
RUN go mod tidy
RUN GOOS=${GOOS} GOARCH=${GOARCH} go build -o /out/example main.go

FROM scratch
COPY --from=build /out/ /

#FROM scratch AS bin-windows
#COPY --from=build /out/example /example.exe

