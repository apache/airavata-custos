//
// Created by jens on 18/08/2021.
//

#ifndef __PAM_OAUTH2_DEVICE_PAM_OAUTH2_LOG_HPP
#define __PAM_OAUTH2_DEVICE_PAM_OAUTH2_LOG_HPP


#include <cstdio>
#include <exception>
#include <stdarg.h>

struct pam_handle;
class BaseError;  // defined in pam_oauth2_excpt


class pam_oauth2_log {
public:
    //! log levels a subset of those of syslog.
    // Note errors cannot be masked.
    enum class log_level_t { DEBUG, INFO, WARN, ERR, OFF };

    //! simple compare against the class' log level
    // constexpr limitations in C++11?
    bool log_this(log_level_t severity) const noexcept;

    //! Constructor for logger
    //! Will accept a null pointer for the PAM handle in which case it logs to stderr (useful for testing, perhaps)
    //! With a PAM handle it logs to syslog via PAM
    pam_oauth2_log(pam_handle *ph, log_level_t lev) noexcept;
    // no copy, but move is OK
    pam_oauth2_log(pam_oauth2_log const &) = delete;
    pam_oauth2_log(pam_oauth2_log &&) = default;
    pam_oauth2_log &operator=(pam_oauth2_log const &) = delete;
    pam_oauth2_log &operator=(pam_oauth2_log &&) = default;
    ~pam_oauth2_log();

    //! Query the log level
    log_level_t log_level() const noexcept { return lev_; }
    //! Change the log level
    // This can't be constexpr in C++11
    void set_log_level(log_level_t logLevel) noexcept { lev_ = logLevel; }
    //! log an exception
    void log(std::exception const &) noexcept;
    //! log one of our exceptions...
    void log(BaseError const &) noexcept;
    //! log a string at a specific level
    //! C stdarg style (not a C++ pack)
    void log(log_level_t, char const *fmt, ...) noexcept;
    //! Return the pam handle
    pam_handle const *get_pam_handle() { return ph_; }

private:
    //! Translation to syslog priority
    // constexpr limitations in C++11?
    int syslog_pri(log_level_t) const noexcept;

    // The PAM handle
    pam_handle *ph_;
    // The current log level, messages at lower level are not reported
    log_level_t lev_;
    // File for non-syslog output, or null
    FILE *log_;
};




#endif //__PAM_OAUTH2_DEVICE_PAM_OAUTH2_LOG_HPP
