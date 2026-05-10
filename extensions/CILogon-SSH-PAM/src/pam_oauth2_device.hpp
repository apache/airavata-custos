#ifndef PAM_OAUTH2_DEVICE_HPP
#define PAM_OAUTH2_DEVICE_HPP

#include <string>
#include <cstdio>
#include "include/pam_oauth2_log.hpp"


/*! @brief userinfo type object (cf RFC 7662)
 */

class Userinfo
{
private:
    std::string sub_,
        username_,
        name_;
    // groups will be sorted alphabetically
    std::vector<std::string> groups_;
public:
    Userinfo(std::string const &sub, std::string const &username, std::string const &name): sub_(sub), username_(username), name_(name) {}

    /*! @brief Add a group to the userinfo groups.
     * Caution: there is no check whether the group is already in the userinfo groups.
     */
    void add_group(std::string const &group);

    /*! @brief Import a vector of group names into the userinfo groups.
     * If there are already groups set, they will be removed (no merge).
     */
    void set_groups(std::vector<std::string> const &groups);

    std::string name() const { return name_; }
    std::string sub() const { return sub_; }
    std::string username() const { return username_; }

    //! Check if a given group is part of the userinfo groups
    bool is_member(std::string const &group) const;

    //! @brief Check whether groups (must be _sorted_; specifed through iterators) have any overlap with the userinfo groups.
    //! False is returned only if they are wholly distinct.
    bool intersects(std::vector<std::string>::const_iterator beg,
		    std::vector<std::string>::const_iterator end) const;
};


// TODO: improve this struct
class DeviceAuthResponse
{
public:
    std::string user_code,
        verification_uri,
        verification_uri_complete,
        device_code;
    std::string get_prompt(const int qr_ecc);
};

void make_authorization_request(Config const &config,
				pam_oauth2_log &logger,
				std::string const &client_id,
                                std::string const &client_secret,
                                std::string const &scope,
                                std::string const &device_endpoint,
                                DeviceAuthResponse *response);

void poll_for_token(Config const &config,
		    pam_oauth2_log &logger,
		    std::string const &client_id,
                    std::string const &client_secret,
                    std::string const &token_endpoint,
                    std::string const &device_code,
                    std::string &token);

Userinfo get_userinfo(Config const &config,
		      pam_oauth2_log &logger,
		      std::string const &userinfo_endpoint,
		      std::string const &token,
		      std::string const &username_attribute);

#endif // PAM_OAUTH2_DEVICE_HPP
