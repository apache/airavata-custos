#ifndef PAM_OAUTH2_DEVICE_LDAPQUERY_H
#define PAM_OAUTH2_DEVICE_LDAPQUERY_H

#define LDAPQUERY_ERROR -1
#define LDAPQUERY_TRUE 1 
#define LDAPQUERY_FALSE 0


#ifdef __cplusplus
extern "C" {
#endif

// Translating the C++ enum class defined in pam_oauth2_log into C
enum ldap_loglevel_t { LDAP_LOGLEVEL_DEBUG, LDAP_LOGLEVEL_INFO, LDAP_LOGLEVEL_WARN, LDAP_LOGLEVEL_ERR, LDAP_LOGLEVEL_OFF };

    
//! Convert a string value to an LDAP scope (which should be ber_int_t)
//! (ie must be one of "base", "subtree", "onelevel", or "children", or a few common variations thereof)
//! On failure, returns -1000
int ldap_scope_value(char const *);

//! Query an LDAP server for attr, which should match value (case sensitively)
//! pamh is used for logging
int ldap_check_attr(void const *pamh, enum ldap_loglevel_t log,
		    const char *host, const char *basedn, int scope,
                    const char *user, const char *passwd,
                    const char *filter, const char *attr,
                    const char *value);

//! Run a query against an LDAP server and see whether it is successful
//! pamh is used for logging
int ldap_bool_query(void const *pamh, enum ldap_loglevel_t log,
		    const char *host, const char *basedn, int scope,
		    const char *user, const char *passwd,
		    const char *query);


#ifdef __cplusplus
}
#endif

#endif  // PAM_OAUTH2_DEVICE_LDAPQUERY_H
