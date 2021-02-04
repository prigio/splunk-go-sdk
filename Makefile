#Environment settings for cross compilation
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
ENV_OSX=-e GOOS=darwin -e GOARCH=amd64
ENV_WIN=-e GOOS=windows -e GOARCH=amd64
ENV_LIN=-e GOOS=linux -e GOARCH=amd64
#this is there the src files are located, within the container
#the name of the directory might be used by GO for the name of the executable
#this is where build files are to be stored, within the container
#BUILDSDIR=/usr/local/bin
#VOL_SRC="${PWD}/src:${WORKDIR}"
#VOL_BUILDS="${PWD}/builds:${BUILDSDIR}"
#the libraries here are populated by the go container itself
#VOL_LIBS="${PWD}/go_build_libs:/go"

default: build




build:
	docker build --target bin --output bin/ --platform local .

