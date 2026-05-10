#ifndef PAM_OAUTH2_DEVICE_METADATA_HPP
#define PAM_OAUTH2_DEVICE_METADATA_HPP

#include <string>

class Metadata
{
public:
    void load(const char *path);
    std::string project_id;
};

#endif // PAM_OAUTH2_DEVICE_METADATA_HPP