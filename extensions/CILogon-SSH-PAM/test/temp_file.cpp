//
// Created by jens.jen@stfc.ac.uk on 25/06/2021.
// Fairly Unix specific, but this was just designed to be a simple RAII file
//
// TODO tidy up all the different errors and exceptions
// TODO handle case where a named file overwrites an existing file
// TODO check for thread safety
// TODO allocating FILENAME_MAX for every instance is a bit wasteful...

#include "temp_file.hpp"

#include <cstdlib>
#include <cstring>
#include <unistd.h>
#include <exception>
//#include <algorithm>


// As it says on the label. Returns true if successful.
static bool write_data_to_file(FILE *fp, char const *data) noexcept;

// very C-ish
static void make_temp( char *filename, size_t size, FILE **file )
{
    constexpr char const *tempname = "/tmp/pam_oauth2_XXXXXX"; // must have six Xs
    if(strlen(tempname) >= size)
	throw "Really can't happen";
    strncpy(filename, tempname, size);
    int fd = mkstemp(filename);
    if(fd < 0)
	throw "Failed to create temp file";
    *file = fdopen(fd, "w");  // foo inherits the file descriptor
    if(!*file)
	throw "Failed to create file object (out of memory?)";
}


static void save_cwd( char *filename, size_t size )
{
    if (!getcwd(filename, size))
	throw "Insufficient memory"; // can't happen
}



static void save_filename(char const *filename, char *dest, size_t size)
{
    char *p = dest;
    if(*filename != '/') {
	// For a relative path, save the current directory path (in which the relative path is created)
	// (unless another thread changes it between the getcwd and the file creation?!)
	save_cwd( dest, size );
	size_t len = strlen(p);
	// Add '/' at the end of the path, so we can concatenate the filename
	dest[len] = '/';
	dest[++len] = '\0';
	p += len;
	size -= len;
    }
    if (strlen(filename) >= size)
	throw "Filename too long";
    strncpy(p, filename, size);
}





TempFile::TempFile(const char *contents)
{
    FILE *foo;
    make_temp(fname_, sizeof(fname_), &foo);
    if(!write_data_to_file(foo, contents)) {
	fclose(foo);
	unlink(fname_);
	throw "Unable to write data to file";
    }
    fclose(foo);
}



// We don't have stringview or filesystem until C++17
TempFile::TempFile(std::string const &s)
{
    FILE *foo;
    make_temp(fname_, sizeof(fname_), &foo);
    bool did_it_work_eh = write_data_to_file(foo, s.c_str());
    fclose(foo);
    if(!did_it_work_eh) {
        unlink(fname_);
        throw "Failed to write data to file";
    }
}




TempFile::TempFile(char const *filename, std::string const &contents)
{
    save_filename(filename, fname_, sizeof(fname_));
    FILE *foo = fopen(fname_, "w");
    if(!foo)
        throw "Failed to open file for writing";
    bool did_it_work_eh = write_data_to_file(foo, contents.c_str());
    fclose(foo);
    if(!did_it_work_eh) {
	unlink(fname_);
	throw "Failed to write data to file";
    }
}



TempFile::TempFile(const char *filename, const char *contents)
{
    save_filename(filename, fname_, sizeof(fname_));
    FILE *f = fopen(filename, "w");
    if(!f)
        throw "failed to create file for writing";
    if(!write_data_to_file(f, contents)) {
	fclose(f);
	unlink(filename);
	throw "Failed to write data to file";
    }
    fclose(f);
}


TempFile::~TempFile()
{
    // RAII
    unlink(fname_);
}


std::string
TempFile::filename() const
{
    return std::string{fname_};
}


// TODO CWD could change for relative named files
std::string
TempFile::dirname() const
{
    char const *p = fname_, *q = strrchr(fname_, '/');
    if(q) {
        std::string path(fname_, q-p);
        return path;
    }
    // can't happen: path wasn't set, so return CWD
    char buf[FILENAME_MAX];
    if(!getcwd(buf, FILENAME_MAX))
        throw std::bad_alloc();  // can't happen
    return  std::string(buf);
}


bool
write_data_to_file(FILE *fp, char const *data) noexcept
{
    size_t len = strlen(data);
    return len == fwrite(data, 1, len, fp);
}
