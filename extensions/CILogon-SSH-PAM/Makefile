CXXFLAGS=-Wall -fPIC -std=c++11
CFLAGS=-Wall -fPIC

LDLIBS=-lpam -lcurl -lldap -llber

objects = src/pam_oauth2_device.o \
		  src/include/config.o \
          src/include/metadata.o \
		  src/include/pam_oauth2_curl.o \
		  src/include/pam_oauth2_log.o \
		  src/include/ldapquery.o \
		  src/include/nayuki/BitBuffer.o \
		  src/include/nayuki/QrCode.o \
		  src/include/nayuki/QrSegment.o

all: pam_oauth2_device.so

%.o: %.c %.h
	$(CC) $(CFLAGS) -c $< -o $@

%.o: %.cpp %.hpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

pam_oauth2_device.so: $(objects)
	$(CXX) -shared $^ $(LDLIBS) -o $@

.PHONY: clean distclean install
clean:
	rm -f $(objects)

distclean: clean
	rm -f pam_oauth2_device.so

install: pam_oauth2_device.so
	install -D -t $(DESTDIR)$(PREFIX)/lib/security pam_oauth2_device.so
	install -m 600 -D config_template.json $(DESTDIR)$(PREFIX)/etc/pam_oauth2_device/config.json

.PHONY: test tests

test tests:
	make -C test
