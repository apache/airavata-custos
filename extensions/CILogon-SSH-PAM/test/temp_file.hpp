//
// Created by jens.jensen@stfc.ac.uk on 25/06/2021.
//
// This class uses RAII to create a temporary file with a specific content.
// This is used to create files for testing which are automatically cleared after the test has finished.
// It really should use boost or something for portable file handling but it needs to minimise dependencies


#ifndef __PAM_OAUTH2_DEVICE_TEMP_FILE_HPP
#define __PAM_OAUTH2_DEVICE_TEMP_FILE_HPP

#include <string>
#include <cstdio>

/** \brief Create a temporary file with specified content using RAII */

class TempFile {
private:
    char fname_[FILENAME_MAX];
public:
    /** @brief construct file with given contents */
    TempFile(std::string const &contents);
    /** @brief construct file with given contents */
    TempFile(char const *contents);
    /** @brief construct file with given name and contents */
    TempFile(char const *filename, char const *contents);
    /** @brief construct file with given name and contents */
    TempFile(char const *filename, std::string const &contents);

    TempFile(TempFile const &) = delete;
    TempFile(TempFile &&) = delete;
    TempFile operator=(TempFile const &) = delete;
    TempFile &operator=(TempFile &&) = delete;
    ~TempFile();

    /** Return the full path/name of the file */
    std::string filename() const;
    /** Return the directory of the file */
    std::string dirname() const;
};


#endif //__PAM_OAUTH2_DEVICE_TEMP_FILE_HPP
