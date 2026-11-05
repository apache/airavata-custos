#include <cstdio>
#include <fstream>
#include "gtest/gtest.h"
#include "include/config.hpp"
#include "include/nlohmann/json.hpp"

#define CLIENT_ID "client_id"

using json = nlohmann::json;

namespace
{

TEST(ConfigTest, MissingFile)
{
    Config config;
    ASSERT_THROW(config.load("data/missing.json"), std::runtime_error);
}

TEST(ConfigTest, WrongFormat)
{
    Config config;
    ASSERT_THROW(config.load("data/template_wrong.json"), json::parse_error);
}

TEST(ConfigTest, Empty)
{
    Config config;
    ASSERT_THROW(config.load("data/template_empty.json"), json::out_of_range);
}

TEST(ConfigTest, NoLdap)
{
    Config config;
    config.load("data/template_noldap.json");
    EXPECT_EQ(config.client_id, CLIENT_ID);
    EXPECT_TRUE(config.ldap_host.empty());
}

TEST(ConfigTest, Full)
{
    Config config;
    config.load("../config_template.json");
    EXPECT_EQ(config.client_id, CLIENT_ID);
    EXPECT_EQ(config.ldap_host, "ldaps://ldap-server:636");
    EXPECT_EQ(config.usermap["*bypass*"].count("root"), 1);
    EXPECT_EQ(config.usermap["provider_user_id_1"].count("bob"), 1);
    EXPECT_EQ(config.usermap.size(), 3);
    EXPECT_EQ(config.qr_error_correction_level, 0);
}

TEST(ConfigTest, IncompleteLdap)
{
    for (auto const *ldap : {R"j({"host":"ldap://h","attr":"uid"})j", R"j({"host":"ldap://h","filter":"(uid=%s)"})j", R"j({"preauth":"(uid=%s)"})j"})
    {
        json j = json::parse(std::ifstream("data/template_noldap.json"));
        j["ldap"] = json::parse(ldap);
        std::ofstream("data/ldap.json") << j;
        Config config;
        EXPECT_THROW(config.load("data/ldap.json"), std::runtime_error) << ldap;
    }
    std::remove("data/ldap.json");
}

} // namespace