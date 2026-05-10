//
// Created by jens on 26/07/2021.
// This is a separate header file to make it easier to regression test the implementation
//

#ifndef __PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_IMPL_HPP
#define __PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_IMPL_HPP

#include "pam_oauth2_curl.hpp"
#include <curl/curl.h>


// namespace pam_oauth2_curl_impl {

//@brief callback for curl
size_t WriteCallback(char const *contents, size_t size, size_t nmemb, void *userp);


//! @brief small structure for handling aux metadata (not stored in CURL handle) for each POST or GET call
struct call_data {
    std::string callback_data;
    std::string auz_hdr;
    pam_oauth2_curl::params post_data;
    curl_slist *headers;
    pam_oauth2_curl::credential cred;
    // Classic C string error buffer
    char errbuf[CURL_ERROR_SIZE];

    call_data() : callback_data(), auz_hdr(), post_data(), headers(nullptr), cred() { }
    ~call_data()
    {
        if(headers) curl_slist_free_all(headers);
    }
};



struct pam_oauth2_curl_impl {
    CURL *curl;
    CURLcode ret;
    std::vector<call_data> calls;
    static std::string make_post_data(pam_oauth2_curl::params const &data);
    // RFC 3986 section 2.2 (and 2.4 for '%').
    static std::string reserved;
    // from curl.h; the enum id of the SSL backend that we are using
    // curlssl_backend backend;

public:
    pam_oauth2_curl_impl(Config const &config);
    ~pam_oauth2_curl_impl();
    //@ reset curl handle to only have basic shared options
    void reset(Config const &config);
    //@ add callback data, linking the curl handle to a particular call_data (which are specific to each call)
    void add_call_data(call_data &);
    //@ Add a credential to the current curl options/header
    void add_credential(call_data &data, pam_oauth2_curl::credential &&cred);
    //@ RFC3986 encode
    std::string encode(std::string const &in);
    //@ Add parameters to a parameter list.
    // This needs to be here so we can call it from add_credential because it is no longer static
    pam_oauth2_curl::params &add_params(pam_oauth2_curl::params &params, std::string const &key, std::string const &value);
    //@ return true if string contains a reserved character
    static bool contains_reserved(std::string const &);
};


//} // namespace pam_oauth2_curl_impl

#endif //__PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_IMPL_HPP
