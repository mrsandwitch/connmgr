all: ConnMgr
ConnMgr:
	go build connmgr

install:
	cp connmgr /usr/local/bin/ConnMgr
