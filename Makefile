include $(GOROOT)/src/Make.inc

TARG=bitbucket.org/zombiezen/ftp
GOFILES=\
	client.go\
	doc.go\
	reply.go\

include $(GOROOT)/src/Make.pkg
