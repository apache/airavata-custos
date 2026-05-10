# Assume Google Test is installed on the system
# GTEST_DIR = ./googletest-release-1.8.1/googletest
GTEST_DIR = /usr

SRC_DIR = ../src

CPPFLAGS += -isystem -I$(GTEST_DIR)/include -I$(SRC_DIR)/include -I$(SRC_DIR)

CXXFLAGS += -g -Wall -Wextra -Wno-unused-parameter -pthread -std=c++11

LDLIBS=-lpam -lcurl -lldap -llber

TESTS = test_config unit test_pam_oauth2_device

GTEST_HEADERS = $(GTEST_DIR)/include/gtest/*.h \
                $(GTEST_DIR)/include/gtest/internal/*.h

objects = $(SRC_DIR)/pam_oauth2_device.o \
		  $(SRC_DIR)/include/config.o \
		  $(SRC_DIR)/include/ldapquery.o \
		  $(SRC_DIR)/include/metadata.o \
		  $(SRC_DIR)/include/nayuki/BitBuffer.o \
		  $(SRC_DIR)/include/nayuki/QrCode.o \
		  $(SRC_DIR)/include/nayuki/QrSegment.o \
		  $(SRC_DIR)/include/pam_oauth2_curl.o \
		  $(SRC_DIR)/include/pam_oauth2_log.o

all: $(TESTS)
	for test in $(TESTS); do ./$${test}; done


.PHONY: clean distclean

clean:
	rm -f gtest.a gtest_main.a *.o $(objects) unit.o temp_file.o test_pam_oauth2_device.o

distclean: clean
	rm -f $(TESTS)

# System libraries for gtest (includes the default main function)
GTEST_LIBS = -lgtest_main -lgtest -lpthread

%.o: %.c %.h
	$(CXX) $(CXXFLAGS) -c $< -o $@

test_config.o: test_config.cpp $(GTEST_HEADERS) $(SRC_DIR)/include/config.hpp
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -I$(SRC_DIR) -c test_config.cpp

test_config: test_config.o $(SRC_DIR)/include/config.o
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) $(GTEST_LIBS) $^ -o $@

unit: unit.o temp_file.o $(objects)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) $^  $(LDLIBS) $(GTEST_LIBS) -o $@

test_pam_oauth2_device.o: test_pam_oauth2_device.cpp $(GTEST_HEADERS) $(SRC_DIR)/include/config.hpp $(SRC_DIR)/pam_oauth2_device.hpp
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -I$(SRC_DIR) $(LDLIBS) -c test_pam_oauth2_device.cpp

test_pam_oauth2_device: test_pam_oauth2_device.o $(objects)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) $(GTEST_LIBS)  $^ $(LDLIBS) -o $@

# unit tests, not part of the standard test suite (which tests pam_oauth_device.o itself)
test_pam_oauth2_curl: test_pam_oauth2_curl.cpp  ../src/include/pam_oauth2_curl.o
	$(CXX) $(CPPFLAGS) $(CXXFLAGS)  -o $@ -L $(SRC_DIR)/include $^ $(GTEST_LIBS) -lcurl

