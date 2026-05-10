/************************************************************
 * *** Unit and correctness testing for pam_oauth2_device
 * Normally we test the public API but in this case we need to test the private API as well
 * jens.jensen@stfc.ac.uk
 */

#include <gtest/gtest.h>
#include <vector>
#include <string>
#include <security/pam_appl.h>
#include "config.hpp"
#include "metadata.hpp"
#include "pam_oauth2_device.hpp"
#include "temp_file.hpp"
#include <fstream>
#include <algorithm>
#include <iterator>
#include <cstdlib>


/** \brief Check whether the contents of a file matches exactly that of the string being passed to it
 * \param filename - the name of the file to be scanned relative to CWD
 * \param string - the string to compare
 * @return -1 if matching, -1000 if file is absent; or location of first mismatch if not matching
 */
ssize_t cmp_file_string(char const *filename, std::string const &string);

enum class ConfigSection { TEST_CLOUD, TEST_GROUP, TEST_USERMAP, TEST_LDAP };

/** \brief Make a dummy Config class for testing */
Config make_dummy_config(ConfigSection, Userinfo const &);

/** \brief make a dummy userinfo class */
Userinfo make_dummy_userinfo(std::string const &username);

/** Test function for cloud section of is_authorized() */
bool is_authorized_cloud(Userinfo &ui, std::string const &username_local, std::vector<std::string> const &groups);

/** Test function for group section of is_authorized() */
bool is_authorized_group(Userinfo &ui, std::string const &username_local, std::string const &service_name, std::vector<std::string> const &groups);

/** Test function for the local usermap section of is_authorized (mapping remote usernames to authorised local usernames */
bool is_authorized_local(Userinfo &ui, std::string const &username_local);

/* copied prototypes for private (compilation unit) functions from pam_oauth2_device.cpp */
std::string getQr(const char *text, const int ecc = 0, const int border = 1);

class DeviceAuthResponse;

void make_authorization_request(Config const &,
				pam_oauth2_log &,
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

void show_prompt(pam_handle_t *pamh,
		 int qr_error_correction_level,
		 DeviceAuthResponse *device_auth_response);

bool is_authorized(Config const &config,
		   pam_oauth2_log &logger,
		   std::string const &username_local,
		   Userinfo const &userinfo,
		   char const *metadata_path = nullptr);


TEST(PamOAuth2Unit, QrCodeTest)
{
    char const *text = "I want to think audibly this evening. I do not want to make a speech and if you find me this evening speaking without reserve, pray, consider"
		       " that you are only sharing the thoughts of a man who allows himself to think audibly, and if you think that I seem to transgress the limits"
		       " that courtesy imposes upon me, pardon me for the liberty I may be taking.";
    char const *loremipsum = "loremipsum";
EXPECT_EQ(cmp_file_string("data/qr1.0.txt", getQr(loremipsum, 0, 1)), -1);
EXPECT_EQ(cmp_file_string("data/qr1.1.txt", getQr(loremipsum, 1, 1)), -1);
EXPECT_EQ(cmp_file_string("data/qr1.2.txt", getQr(loremipsum, 2, 1)), -1);
EXPECT_EQ(cmp_file_string("data/qr2.0.txt", getQr(text, 0, 1)), -1);
EXPECT_EQ(cmp_file_string("data/qr2.1.txt", getQr(text, 1, 1)), -1);
EXPECT_EQ(cmp_file_string("data/qr2.2.txt", getQr(text, 2, 1)), -1);
}

TEST(PamOAuth2Unit, IsAuthorized)
{
    // Userinfo contains the remote username. Note the suffix is (or will be) configured as ".test" to match a local username "fred"
    Userinfo ui{make_dummy_userinfo("fred.test")};
    // groups denotes the groups assigned to the project id
    std::vector<std::string> groups;
    // No or wrong groups, right username
EXPECT_TRUE( !is_authorized_cloud(ui, "fred", groups));
groups.push_back("sknamp");
EXPECT_TRUE( !is_authorized_cloud(ui, "fred", groups));
// Groups, correct username
groups.push_back("bleps");
groups.push_back("plamf");
// Check groups is sorted
std::sort(groups.begin(), groups.end());
// One group is right, username is right
EXPECT_TRUE( is_authorized_cloud(ui, "fred", groups));
// Right groups, wrong username
EXPECT_TRUE( !is_authorized_cloud(ui, "barney", groups));
// Now for the groups test, starting with a service name which is not one of fred's groups
EXPECT_TRUE(!is_authorized_group(ui, "fred", "bylzp", groups));
// service name is one of fred's Userinfo groups but not in groups, and but username is different
EXPECT_TRUE(!is_authorized_group(ui, "wilma", "plempf", groups));
// service name is one of fred's Userinfo groups but not in project_id groups
EXPECT_TRUE(is_authorized_group(ui, "fred", "plempf", groups));
// service name is in project_id groups but not in fred's Userinfo groups
EXPECT_TRUE(!is_authorized_group(ui, "fred", "plamf", groups));
// local map test: remote name is in the list but local name doesn't match
EXPECT_TRUE(!is_authorized_local(ui, "gnumpf"));
// local map test: remote name is in the list and a local name matches
EXPECT_TRUE(is_authorized_local(ui, "fred"));
// local map test: remote name is not in list
Userinfo ui2{"0123456789abcdef", "barney.test", "barney"};
EXPECT_TRUE(!is_authorized_local(ui2, "barney"));
}


ssize_t
cmp_file_string(char const *filename, std::string const &string)
{
    std::ifstream foo(filename, std::ios_base::binary);
    if(!foo)
        return -1000;
    ssize_t index = 0;
    // istream_iterator doesn't work because it parses the input
    auto p = string.cbegin();
    auto const q = string.cend();
    while(p != q) {
        // get returns EOF at, er, EOF, and EOF is never a character
        char c1, c2;
        c1 = foo.get();
        c2 = *p++;
        if(c1 != c2) {
            return index;
        }
    }
    return -1;    // match
}



Config
make_dummy_config(ConfigSection section, Userinfo const &ui)
{
    Config cf;
    // All members are public! and have no non-default initialisers
    // Boolean selectors of test section
    cf.cloud_access = cf.group_access = false;
    cf.local_username_suffix = ".test";
    switch (section) {
	case ConfigSection::TEST_CLOUD:
	    cf.cloud_access = true;
	    // The following three variables are needed: cloud_username, local_username_suffix, cloud_endpoint
	    // "cloud username" is the remote username
	    cf.cloud_username = "fred.test";
	    // endpoint is set later as we don't know it yet
	    // metadata_file is set later as we don't know it yet
	    break;
	case ConfigSection::TEST_GROUP:
	    cf.group_access = true;
	    break;
	case ConfigSection::TEST_LDAP:
	    break;
	case ConfigSection::TEST_USERMAP:
	    // create a dummy usermap to test against
	    std::set<std::string> fred{"fred", "blips", "flopsy"};
	    std::set<std::string> wilma{"wilma", "betty", "blaps"};
	    cf.usermap.insert(std::pair<std::string,std::set<std::string>>{"fred.test", fred});
	    cf.usermap.insert(std::pair<std::string,std::set<std::string>>{"wilma.test", wilma});
	    break;
	// no default
    }
    return cf;
}


Userinfo
make_dummy_userinfo(std::string const &username)
{
    Userinfo ui{"0123456789abcdef", username, "jdoe"};
    // Note groups are not added alphabetically
    ui.add_group("splomp");
    ui.add_group("plempf");
    ui.add_group("bleps");
    return ui;
}



Metadata
make_dummy_metadata()
{
    Metadata md;
    // This is currently a public member! but will not test the load function
    md.project_id = "iristest";
    return md;
}


std::string
make_groups_json(std::vector<std::string> const &groups)
{
    // Slightly hacky JSON construction
    std::string contents{"{\"groups\":[\""};
    if(!groups.empty()) {
	auto end = groups.cend()-1;
	std::for_each(groups.cbegin(), end, [&contents](std::string const &grp) { contents += grp; contents += "\",\""; });
	contents += *end;
    }
    contents += "\"]}";
    return contents;
}


bool
is_authorized_cloud(Userinfo &ui, std::string const &username_local, std::vector<std::string> const &groups)
{
    Config cf{make_dummy_config(ConfigSection::TEST_CLOUD, ui)};
    pam_oauth2_log log(nullptr, pam_oauth2_log::log_level_t::DEBUG);
    TempFile metadata("{\"project_id\":\"iristest\"}");
    cf.metadata_file = metadata.filename();
    // The project id is the name of the file
    TempFile cloud( "iristest", make_groups_json(groups));
    // curl can read a local file!
    cf.cloud_endpoint = "file://" +  cloud.dirname();
    // The project id file should be passed in with the config.
    // Destructors are not called until the call has returned (so c_str()s are safe)
    return is_authorized(cf, log, username_local, ui);
}



bool
is_authorized_group(Userinfo &ui, std::string const &username_local, std::string const &service_name, std::vector<std::string> const &groups)
{
    Config cf{make_dummy_config(ConfigSection::TEST_GROUP, ui)};
    pam_oauth2_log log(nullptr, pam_oauth2_log::log_level_t::DEBUG);
    cf.group_service_name = service_name;     // gets copied (string constructor)
    return is_authorized(cf, log, username_local, ui, nullptr /* only needed for cloud */);
}



bool
is_authorized_local(Userinfo &ui, std::string const &username_local)
{
    Config cf{make_dummy_config(ConfigSection::TEST_USERMAP, ui)};
    pam_oauth2_log log(nullptr, pam_oauth2_log::log_level_t::DEBUG);
    return is_authorized(cf, log, username_local, ui, nullptr);
}
