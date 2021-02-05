#FROM --platform=${BUILDPLATFORM} golang:1.16-rc AS build
FROM golang:1.16-rc AS build
WORKDIR /src
#ENV CGO_ENABLED=0
COPY examples/hello .
ARG TARGETOS
ARG TARGETARCH
# for gitlab private repos, authentication with tokens is necessary 
# and GOPRIVATE must skip verification for git.cocus.com
RUN echo "machine git.cocus.com login go-build password 6H93rHDDn-HnL26tTtRP" > ~/.netrc
RUN ls 
RUN GOPRIVATE=git.cocus.com/* GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/example main.go

FROM scratch AS bin-unix
COPY --from=build /out/example /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

#FROM scratch AS bin-windows
#COPY --from=build /out/example /example.exe

FROM bin-${TARGETOS} AS bin
