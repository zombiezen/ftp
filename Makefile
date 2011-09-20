include $(GOROOT)/src/Make.inc

TARG=bitbucket.org/zombiezen/ftp
GOFILES=\
	client.go\
	ftp.go\

include $(GOROOT)/src/Make.pkg
