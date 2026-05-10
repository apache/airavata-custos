//
// Created by jens on 23/07/2021.
// PIMPL/RAII abstraction of the curl library for pam_oauth2_device
// for synchronous HTTPS GET and POST as required by pam_oauth2_device
//
// NOTES:
// 1. Functions can throw exceptions which are currently defined in pam_oauth2_device.hpp
// 2. Not thread safe since curl_global_init is not thread safe, and the calls use the same curl handle throughout
//

#ifndef __PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_HPP
#define __PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_HPP

#include <string>
#include <vector>
#include <utility>
#include <memory>


// pimpl (defined in pam_oauth2_curl_impl.hpp and implemented in pam_oauth2_curl.cpp)
class pam_oauth2_curl_impl;
// declared in pam_oauth2_log.hpp
// class pam_oauth2_log;
// defined in config.hpp
class Config;



class pam_oauth2_curl {
private:
    // A unique pointer needs to know the size of the pointee which goes against the logic of the pimpl
    pam_oauth2_curl_impl *impl_;
//    pam_oauth2_log &log_;
public:
    pam_oauth2_curl(Config const &config);
    //    pam_oauth2_curl(Config const &config, pam_oauth2_log &logger);
    ~pam_oauth2_curl();
    pam_oauth2_curl(pam_oauth2_curl const &) = delete;
    pam_oauth2_curl(pam_oauth2_curl &&) = delete;
    pam_oauth2_curl &operator=(pam_oauth2_curl const &) = delete;
    pam_oauth2_curl &operator=(pam_oauth2_curl &&) = delete;

    //! parameter list; use add_params to add stuff to it (caller should treat it as an opaque type)
    using params = std::vector<std::pair<std::string,std::string>>;

    // Need separate credentials to accommodate CHTC patches.
    // The reference is RFC 6749 section 2.3.1 (and RFC2617)
    // This would have been a std::variant perhaps in later versions of C++
    class credential {
    protected:
        /** @brief Different types of credentials
         - DENIED is always a failed authentication, can be used as default
         - NONE is no authentication, server has to reject if it doesn't like it
         - BASIC is RFC2617 username and password, OAuth2 style
         - TOKEN is bearer token
         - SECRET is sending the client_id/secret (as NOT RECOMMENDED by RFC6749 section 2.3.1...)
         */
        enum class type { DENIED, NONE, BASIC, TOKEN, SECRET } type_;
        std::string un_, pw_;
        std::string token_;
    public:
        credential() : type_(credential::type::NONE), un_(), pw_(), token_() { }
        credential(std::string const &username, std::string const &password) : type_(credential::type::BASIC), un_(username), pw_(password), token_() { }
        credential(std::string const &token) : type_(credential::type::TOKEN), un_(), pw_(), token_(token) { }
	// The int is just there to disambiguate the overload, in lieu of something cleverer like a factory
        credential(std::string const &client_id, std::string const &client_secret, int) : type_(credential::type::SECRET), un_(client_id), pw_(client_secret), token_() { }
        // don't copy if possible
	credential(credential const &) = delete;
	credential &operator=(credential const &) = delete;
	// move only
	credential(credential &&) = default;
	credential &operator=(credential &&) = default;
	~credential();
	friend class pam_oauth2_curl_impl;
    };

    //! @brief Credential factory kind of thing
    credential make_credential(Config const &);

    //! perform a HTTP GET or POST synchronously, returning result
    std::string call(Config const &config, std::string const &url);
    //! perform a HTTP POST synchronously, returning result
    std::string call(Config const &config, std::string const &url, params const &postdata);
    //! perform a HTTP call with custom credentials
    std::string call(Config const &config, std::string const &url, credential &&cred);
    //! add parameters to parameter list
    params &add_params(params &params, std::string const &key, std::string const &value);
    //! URL encode.
    std::string encode(std::string const &);
};


#endif //__PAM_OAUTH2_DEVICE_PAM_OAUTH2_CURL_HPP
