PROC_NAME = echogy
RELEASE_PATH = release
PACKAGE_PATH = release
SERVER_PATH = cmd

install:
	@go get

build-cgo:
	cd ${SERVER_PATH} &&  GOOS=linux GOARCH=amd64 go build -race -ldflags "-s -w" -o $(RELEASE_PATH)/webssh

build:
	cd ${SERVER_PATH} &&  GOOS=linux GOARCH=amd64 go build -o $(RELEASE_PATH)/webssh

clean:
	# clean package
	rm -rf ${PACKAGE_PATH}
	# clean server build
	rm -rf ${SERVER_PATH}/${RELEASE_PATH}

package: clean build
	mkdir -p ${PACKAGE_PATH}
	cp ${SERVER_PATH}/${RELEASE_PATH}/webssh  ${PACKAGE_PATH}
	cp ${SERVER_PATH}/config.example.json  ${PACKAGE_PATH}/config.json