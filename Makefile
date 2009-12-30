include $(GOROOT)/src/Make.$(GOARCH)

TARG=mypackage
GOFILES=\
	memcache.go\

include $(GOROOT)/src/Make.pkg

