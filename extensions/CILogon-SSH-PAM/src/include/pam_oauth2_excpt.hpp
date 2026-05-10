//
// Created by jens on 27/07/2021.
//
// Exceptions and logging - definitions

#ifndef __PAM_OAUTH2_DEVICE_PAM_OAUTH2_EXCPT_HPP
#define __PAM_OAUTH2_DEVICE_PAM_OAUTH2_EXCPT_HPP


#include <exception>
#include <string>
// How portable is this?
#include <security/_pam_types.h>
#include "pam_oauth2_log.hpp"


class BaseError : public std::exception
{
    // TODO temporary solution?
    std::string msg_, details_;
    // Severity level to log this exception at
    pam_oauth2_log::log_level_t severity_;
    // The logger is our friend
    friend class pam_oauth2_log;
public:
    BaseError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : msg_(msg), details_(), severity_(severity) { }

    char const *what() const noexcept override { return msg_.c_str(); }

    // Disable copy
    BaseError(BaseError const &) = delete;
    BaseError &operator=(BaseError const &) = delete;
    // Allow moves
    BaseError(BaseError &&) = default;
    BaseError &operator=(BaseError &&) = default;

    //! Optionally add details (like debug info) for longer messages
    void add_details(std::string &&details) { details_ = std::move(details); }
    //! Return the details
    std::string const &details() const noexcept { return details_; }

    //! Return a four character string with the name (or near enough) of the class
    virtual char const *type() const noexcept { return "BASE"; }
    //! If this gives rise to a PAM error, what is it?
    virtual int pam_error() const noexcept { return PAM_AUTH_ERR; }
};


struct ConfigError : public BaseError
{
    ConfigError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : BaseError(msg, severity) { }
    char const *type() const noexcept override { return "CONF"; }
};


struct PamError : public BaseError
{
    PamError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : BaseError(msg, severity) { }
    char const *type() const noexcept override { return "PAM "; }
    int pam_error() const noexcept override { return PAM_SYSTEM_ERR; }
};

struct NetworkError : public BaseError
{
    NetworkError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : BaseError(msg, severity) { }
    char const *type() const noexcept override { return "NETW"; }
};

struct TimeoutError : public NetworkError
{
    TimeoutError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : NetworkError(msg, severity) { }
    char const *type() const noexcept override { return "TIME"; }
};

struct ResponseError : public NetworkError
{
    ResponseError(char const *msg, pam_oauth2_log::log_level_t severity = pam_oauth2_log::log_level_t::ERR) : NetworkError(msg, severity) { }
    char const *type() const noexcept override { return "RESP"; }
};


#endif //__PAM_OAUTH2_DEVICE_PAM_OAUTH2_EXCPT_HPP
