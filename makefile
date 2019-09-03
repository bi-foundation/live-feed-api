
PROTOS = EventRouter

all: PROTOS

generate:
    $(MAKE) -C $(PROTOS) clean

clean:
	$(MAKE) -C $(PROTOS) clean

.PHONY:
	clean
	all