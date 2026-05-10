#ifndef PAM_OAUTH2_DEVICE_CONFIG_HPP
#define PAM_OAUTH2_DEVICE_CONFIG_HPP

#include <map>
#include <set>
#include <string>

class Config
{
public:
    void load(const char *path);
    std::string client_id,
        client_secret,
        scope,
        device_endpoint,
        token_endpoint,
        userinfo_endpoint,
        username_attribute,
        ldap_host,
        ldap_basedn,
	ldap_scope,
        ldap_user,
        ldap_passwd,
	ldap_preauth,
        ldap_filter,
        ldap_attr,
	tls_ca_path,
	tls_ca_bundle,
        group_service_name,
        cloud_endpoint,
        cloud_username,
        local_username_suffix,
        metadata_file;
    int qr_error_correction_level;
    bool group_access,
         cloud_access,
         group_and_username_access,
         http_basic_auth,
         client_debug;
    std::map<std::string, std::set<std::string>> usermap;
};

#endif // PAM_OAUTH2_DEVICE_CONFIG_HPP
