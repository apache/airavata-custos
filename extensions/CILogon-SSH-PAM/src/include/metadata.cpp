#include <fstream>

#include "metadata.hpp"
#include "nlohmann/json.hpp"

using json = nlohmann::json;

void Metadata::load(const char *path)
{
    std::ifstream config_fstream(path);
    json j;
    config_fstream >> j;

    project_id = j.at("project_id").get<std::string>();
}