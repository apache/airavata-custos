#include <ldap.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "ldapquery.h"
#include <security/pam_ext.h>
#include <syslog.h>

int ldap_scope_value(char const *string) {
  // Some sort of sensible default for an empty input
  if (!string || !*string)
    return LDAP_SCOPE_SUB;
  if (!strcasecmp(string, "base") || !strcasecmp(string, "baseobject"))
    return LDAP_SCOPE_BASE;
  if (!strcasecmp(string, "one") || !strcasecmp(string, "onelevel"))
    return LDAP_SCOPE_ONE;
  if (!strcasecmp(string, "sub") || !strcasecmp(string, "subtree"))
    return LDAP_SCOPE_SUB;
  // OpenLDAP extension
#ifdef LDAP_SCOPE_SUBORDINATE
  if (!strcasecmp(string, "subordinate") || !strcasecmp(string, "children"))
    return LDAP_SCOPE_SUBORDINATE;
#endif
  return -1000;
}

int ldap_check_attr(void const *pamh, enum ldap_loglevel_t log,
                    const char *host, const char *basedn, int scope,
                    const char *user, const char *passwd, const char *filter,
                    const char *attr, const char *value) {
  LDAP *ld;
  LDAPMessage *res, *msg;
  BerElement *ber;
  BerValue *servercredp;
  char *a, *passwd_local;
  int rc, i;
  struct berval cred;
  struct berval **vals;
  char *attr_local = NULL;
  char *attrs[] = {attr_local, NULL};
  const int ldap_version = LDAP_VERSION3;

  if (ldap_initialize(&ld, host) != LDAP_SUCCESS) {
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP failed init");
    return LDAPQUERY_ERROR;
  }

  if (ldap_set_option(ld, LDAP_OPT_PROTOCOL_VERSION, &ldap_version) != LDAP_SUCCESS) {
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP failed to set v3");
    return LDAPQUERY_ERROR;
  }

  if (user && *user && passwd && *passwd) {
    if (log == LDAP_LOGLEVEL_DEBUG)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_DEBUG, "LDAP set user %s", user);
    passwd_local = (char *)malloc(strlen(passwd) + 1);
    if (!passwd_local) {
      if (log != LDAP_LOGLEVEL_OFF)
        pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP malloc failed");
      return LDAPQUERY_ERROR;
    }
    strcpy(passwd_local, passwd);
    cred.bv_val = passwd_local;
    cred.bv_len = strlen(passwd);
    rc = ldap_sasl_bind_s(ld, user, LDAP_SASL_SIMPLE, &cred, NULL, NULL, &servercredp);
    free(passwd_local);
    if (rc != LDAP_SUCCESS) {
      if (log != LDAP_LOGLEVEL_OFF)
        pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP bind error (search): %s", ldap_err2string(rc));
      return LDAPQUERY_ERROR;
    }
  } else {
    if (log <= LDAP_LOGLEVEL_INFO)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_INFO, "LDAP search anonymous");
    user = NULL; // quick flag to remind ourselves not to unbind
  }

  attr_local = strdup(attr);
  attrs[0] = attr_local;
  if (log == LDAP_LOGLEVEL_DEBUG) {
    pam_syslog(pamh, LOG_AUTHPRIV | LOG_DEBUG,
               "Search base=%s scope=%d filter=%s attr=%s", basedn, scope,
               filter, attrs[0]);
  }
  rc = ldap_search_ext_s(ld, basedn, scope, filter, attrs, 0, NULL, NULL, NULL, 0, &res);
  free(attr_local);
  if (rc != LDAP_SUCCESS) {
    ldap_msgfree(res);
    if (user)
      ldap_unbind_ext_s(ld, NULL, NULL);
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP search for %s failed: %s", attr, ldap_err2string(rc));
    return LDAPQUERY_ERROR;
  }

  rc = LDAPQUERY_FALSE;
  for (msg = ldap_first_message(ld, res); msg != NULL;
       msg = ldap_next_message(ld, msg)) {
    switch (ldap_msgtype(msg)) {
    case LDAP_RES_SEARCH_ENTRY:
      for (a = ldap_first_attribute(ld, res, &ber); a != NULL;
           a = ldap_next_attribute(ld, res, ber)) {
        if ((vals = ldap_get_values_len(ld, res, a)) != NULL) {
          for (i = 0; vals[i] != NULL; ++i) {
            if (strcasecmp(a, attr) == 0) {
              if (strcmp(vals[i]->bv_val, value) == 0) {
                rc = LDAPQUERY_TRUE;
              }
            }
          }
          ldap_value_free_len(vals);
        }
        ldap_memfree(a);
      }
      if (ber != NULL) {
        ber_free(ber, 0);
      }
      break;
    default:
      break;
    }
  }

  ldap_msgfree(res);
  if (user)
    ldap_unbind_ext_s(ld, NULL, NULL);
  return rc;
}

int ldap_bool_query(void const *pamh, enum ldap_loglevel_t log,
                    const char *host, const char *basedn, int scope,
                    const char *user, const char *passwd, const char *query) {
  LDAP *ld;
  LDAPMessage *res, *msg;
  char *passwd_local;
  int rc;
  BerValue *servercredp = NULL;
  struct berval cred;
  char *attrs[] = {LDAP_NO_ATTRS, NULL};
  const int ldap_version = LDAP_VERSION3;

  if (ldap_initialize(&ld, host) != LDAP_SUCCESS) {
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "Failed to init LDAP library");
    return LDAPQUERY_ERROR;
  }

  if (ldap_set_option(ld, LDAP_OPT_PROTOCOL_VERSION, &ldap_version) !=
      LDAP_SUCCESS) {
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "Failed to set version 3");
    return LDAPQUERY_ERROR;
  }

  if (user && *user && passwd && *passwd) {
    if (log == LDAP_LOGLEVEL_DEBUG)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_DEBUG, "Setting LDAP user %s", user);
    passwd_local = (char *)malloc(strlen(passwd) + 1);
    if (!passwd_local) {
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP malloc failed");
      return LDAPQUERY_ERROR;
    }
    strcpy(passwd_local, passwd);
    cred.bv_val = passwd_local;
    cred.bv_len = strlen(passwd);
    rc = ldap_sasl_bind_s(ld, user, LDAP_SASL_SIMPLE, &cred, NULL, NULL, &servercredp);
    free(passwd_local);
    if (rc != LDAP_SUCCESS) {
      if (log != LDAP_LOGLEVEL_OFF)
        pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP bind error (query): %s", ldap_err2string(rc));
      return LDAPQUERY_ERROR;
    }
  } else {
    if (log <= LDAP_LOGLEVEL_INFO)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_INFO, "LDAP query anonymous");
    user = NULL; // quick hack to tell rest of code not to bother to unbind
  }

  if (log == LDAP_LOGLEVEL_DEBUG) {
    pam_syslog(pamh, LOG_AUTHPRIV | LOG_DEBUG, "Query base=%s scope=%d query=%s", basedn, scope, query);
  }
  rc = ldap_search_ext_s(ld, basedn, scope, query, attrs, 1, NULL, NULL, NULL, 0, &res);
  if (rc != LDAP_SUCCESS) {
    ldap_msgfree(res);
    if (user)
      ldap_unbind_ext_s(ld, NULL, NULL);
    if (log != LDAP_LOGLEVEL_OFF)
      pam_syslog(pamh, LOG_AUTHPRIV | LOG_ERR, "LDAP query error: %s", ldap_err2string(rc));
    return LDAPQUERY_ERROR;
  }

  rc = LDAPQUERY_FALSE; // reusing rc
  msg = ldap_first_message(ld, res);
  while (msg) {
    // We are just looking for a search entry, meaning the query returned something
    if (ldap_msgtype(msg) == LDAP_RES_SEARCH_ENTRY)
      rc = LDAPQUERY_TRUE;
    msg = ldap_next_message(ld, msg);
  }
  ldap_msgfree(res);
  if (user)
    ldap_unbind_ext_s(ld, NULL, NULL);
  return rc;
}
