/***** Load the configuration *****/
/* (C) 2022 UKRI-STFC
 * https://www.iris.ac.uk/
 * Jens Jensen, UKRI-STFC
 *
 * Written in C++11 to match the rest of the project
 * https://github.com/stfc/pam_oauth2_device
 */


#include "nlohmann/json.hpp"	// TODO: use the system one
#include <fstream>
#include <iostream>
#include <sstream>
#include <string>
#include <tuple>
#include <forward_list>
#include <stdexcept>
#include <cassert>
#include "config.hpp"


// currently a compile time constant
constexpr bool debug = false;

using json = nlohmann::json;
using list = std::forward_list<std::string>;
using iter = list::const_iterator;


/** Return type
 * @brief we need to return things of different types, and without std::variant (C++17) we need to hack it ourselves
 */
class value final {
public:
    enum class value_type { VT_ERR, VT_NULL, VT_STR, VT_INT, VT_BOOL };

    value() : type_(value_type::VT_NULL), strval_(), intval_(), boolval_() { }

    value(std::string &&str) : type_(value_type::VT_STR), strval_(std::forward<std::string>(str)), intval_(), boolval_() { }

    explicit value(char const *str) : type_(value_type::VT_STR), strval_(str), intval_(), boolval_() { }

    explicit value(int k) : type_(value_type::VT_INT), strval_(), intval_(k), boolval_() { }

    explicit value(bool t) : type_(value_type::VT_BOOL), strval_(), intval_(), boolval_(t) { }

    [[nodiscard]] value_type type() const noexcept { return type_; }

    /** Construct a value based on the JSON structure assuming it is one of those we can encode */
    value(json const &);

    std::string &&get_str() &&
    {
	assert(type_ == value_type::VT_STR);
	return std::forward<std::string>(strval_);
    }
    std::string get_str() const &
    {
	assert(type_ == value_type::VT_STR);
	return strval_;
    }
    int get_int() const
    {
	assert(type_ == value_type::VT_INT);
	return intval_;
    }
    int get_bool() const
    {
	assert(type_ == value_type::VT_BOOL);
	return boolval_;
    }
private:
    value_type type_;
    // No point trying to do this as a union...
    std::string strval_;
    int intval_;
    bool boolval_;

    friend std::ostream &operator<<(std::ostream &, value const &);
    friend bool operator==(value const &, value const &);
    friend bool operator!=(value const &a, value const &b);
};


// Tuple of (1) path, (2) default, (3) required
using item = std::tuple<list,value,bool>;


/** Slightly hacky class for transferring a value to a variable */
class variable final {
private:
    item const place_;
    std::string *s_;
    int *i_;
    bool *b_;
    /** check and transfer value throwing exception if they have different types.
     * As regards const, see comment on set()
     */
    void check_type(value const &, list const &) const;
public:
    variable(list &&path, value &&def, bool reqd, std::string &s) : place_(std::forward<list>(path), std::forward<value>(def), reqd), s_(&s), i_(nullptr), b_(nullptr) {}
    variable(list &&path, value &&def, bool reqd, int &i) : place_(std::forward<list>(path), std::forward<value>(def), reqd), s_(nullptr), i_(&i), b_(nullptr) {}
    variable(list &&path, value &&def, bool reqd, bool &b) : place_(std::forward<list>(path), std::forward<value>(def), reqd), s_(nullptr), i_(nullptr), b_(&b) {}
    /** Set the variable directly from parsing the JSON structure.
     * This is constant from the object's perspective as we just hold a pointer to the target
     */
    void set(json const &) const;
};


/** Read a string attribute
 * @param j initial JSON structure
 * @param next iterator pointing to the current item we're looking for
 * @param end end value for iterator
 */
value load(json j, iter next, iter end, bool required);

/** Read a file, returning its json structure */
json read_file(std::string const &filename);


std::string print_list(list const &);



/**** Principal Load Function ****/


void
Config::load(const char *path)
{
    json const config = read_file(path);
    auto const the_end = config.end();

    // Everything but the 'users' section

    // TODO: this captures the optional attributes but loses the check that if a section is present, some of its attributes are mandatory
    // TODO: It also makes error messages less clear: scope occurs two places and the context may not be passed back
    // TODO: When distributed over several files, it is not clear which file raises the exception
    std::forward_list<variable> const configdata = \
	{ // {path, default, required, target_variable}
	 variable({"oauth","client","id"}, value(""), true, client_id),
	 variable({"oauth","client","secret"}, value(""), true, client_secret),
	 variable({"oauth","scope"},value(""), true, scope),
	 variable({"oauth","device_endpoint"},value(""), true, device_endpoint),
	 variable({"oauth","token_endpoint"},value(""), true, token_endpoint),
	 variable({"oauth","userinfo_endpoint"},value(""), true, userinfo_endpoint),
	 variable({"oauth","username_attribute"},value(""),true, username_attribute),
	 variable({"oauth","local_username_suffix"},value(""),false, local_username_suffix),
	 variable({"qr","error_correction_level"},value(-1),false, qr_error_correction_level),
	 variable({"client_debug"},value(false),false, client_debug),
	 variable({"http_basic_auth"},value(true),false, http_basic_auth),
	 variable({"cloud","access"},value(false),false, cloud_access),
	 variable({"cloud","endpoint"},value(""),false, cloud_endpoint),
	 variable({"cloud","username"},value(""),false, cloud_username),
	 variable({"cloud","metadata_file"},value(""),false, metadata_file),
	 variable({"tls","ca_bundle"},value(""),false, tls_ca_bundle),
	 variable({"tls","ca_path"}, value("/etc/grid-security/certificates"),false, tls_ca_path),
	 variable({"group","access"},value(false),false, group_access),
	 variable({"group","service_name"},value(""),false, group_service_name),
	 variable({"ldap","host"},value(""),false, ldap_host),
	 variable({"ldap","basedn"},value(""),false, ldap_basedn),
	 variable({"ldap","user"},value(""),false, ldap_user),
	 variable({"ldap","passwd"},value(""),false, ldap_passwd),
	 variable({"ldap","scope"},value(""),false, ldap_scope),
	 variable({"ldap","preauth"},value(""),false, ldap_preauth),
	 variable({"ldap","filter"},value(""),false, ldap_filter),
	 variable({"ldap","attr"},value(""),false, ldap_attr)
	};
    for( auto var : configdata )
	var.set(config);

    // The users section is much like before but can be in a separate file, too
    if (config.find("users") != the_end) {
	json j = config.at("users");
	if(j.type() == json::value_t::string) {
	    auto const fn = j.get<std::string>();
	    if(debug)
		std::cerr << "Read users from " << fn << '\n';
	    j = read_file(fn);
	}

        for (auto const &element : j.items())
        {
            for (auto const &local_user : element.value())
            {
                if (usermap.find(element.key()) == usermap.end())
                {
                    std::set<std::string> userset;
                    userset.insert((std::string)local_user);
                    usermap[element.key()] = userset;
		    if(debug)
			std::cerr << "CONF SET users NEW " << element.key() << " " << local_user << std::endl;
                }
                else
                {
                    usermap[element.key()].insert((std::string)local_user);
		    if(debug)
			std::cerr << "CONF SET users ADD " << element.key() << " " << local_user << std::endl;
                }
            }
        }
    }
    else if(debug)
	std::cerr << "No users section found\n";

}




json
read_file(std::string const &filename)
{
    std::ifstream f(filename);
    if(!f.is_open()) {
	std::string computer_says_no = "File not found: ";
	computer_says_no += filename;
	throw std::runtime_error(computer_says_no);
    }
    json fred;
    f >> fred;
    return fred;
}


value
load(json j, iter next, iter end, bool required)
{
    if(debug) {
	if(next != end)
	    std::cerr << "load " << *next << " from " << j << std::endl;
	else
	    std::cerr << "retn " << j << std::endl;
    }
    switch(j.type()) {
    case json::value_t::string:
    {
	std::string val = j.get<std::string>();
	if(next == end)
	    return value(std::move(val));
	// The current entry is a filename
	return load(read_file(val), next, end, required);
    }
    case json::value_t::number_integer:
    case json::value_t::number_unsigned:
    case json::value_t::boolean:
	return value(j);
    case json::value_t::object:
	// We've reached the target but it's an object
	if(next == end)
	    throw "config - reference returns object";
	if(required) {
	    auto node = j.at(*next);
	    return load(node, std::next(next), end, required);
	} else {
	    auto node = j.find(*next);
	    if(node != j.end())
		return load(j.at(*next), std::next(next), end, required);
	    return value();
	}
    default:
	std::ostringstream msg;
	msg << "Unsupported JSON type: " << j;
	throw std::runtime_error(msg.str());
    }
    return value();		// can't happen
}


bool
operator!=(value const &a, value const &b)
{
    return !(a == b);
}


bool
operator==(value const &a, value const &b)
{
    if(a.type_ != b.type_)
	return false;
    switch(a.type_) {
    case value::value_type::VT_ERR:
    case value::value_type::VT_NULL:
	return true;
    case value::value_type::VT_STR:
	return a.strval_ == b.strval_;
    case value::value_type::VT_INT:
	return a.intval_ == b.intval_;
    case value::value_type::VT_BOOL:
	return a.boolval_ == b.boolval_;
    }
    throw std::runtime_error("Unknown value types");	// can't happen
}


std::string
print_list(list const &l)
{
    std::ostringstream os;
    for(auto const s : l)
	os << '/' << s;
    return os.str();
}


value::value(json const &j) : type_(value_type::VT_ERR), strval_(), intval_(), boolval_()
{
    switch(j.type()) {
    case json::value_t::null:
	type_ = value_type::VT_NULL;
	break;
    case json::value_t::boolean:
	type_ = value_type::VT_BOOL;
	boolval_ = j.get<bool>();
	break;
    case json::value_t::string:
	type_ = value_type::VT_STR;
	strval_ = j.get<std::string>();
	break;
    case json::value_t::number_integer:
	type_ = value_type::VT_INT;
	// check for overflow?
	intval_ = j.get<int>();
	break;
    case json::value_t::number_unsigned:
    {
	unsigned int val = j.get<unsigned int>();
	type_ = value_type::VT_INT;
	// check for overflow?
	intval_ = static_cast<int>(val);
	break;
    }
    default:
    {
	std::ostringstream os;
	os << "Cannot construct value from " << j;
	throw std::runtime_error(os.str());
    }
    }
}


std::ostream &
operator<<(std::ostream &os, value const &v)
{
    switch(v.type_) {
    case value::value_type::VT_ERR:
	os << "ERR";
	break;
    case value::value_type::VT_NULL:
	os << "NULL";
	break;
    case value::value_type::VT_STR:
	os << v.strval_;
	break;
    case value::value_type::VT_INT:
	os << v.intval_;
	break;
    case value::value_type::VT_BOOL:
	os << (v.boolval_ ? "true" : "false");
	break;
    }
    return os;
}


void
variable::set(json const &j) const
{
    list const &path(std::get<0>(place_));
    value val = load(j, path.cbegin(), path.cend(), std::get<2>(place_));
    switch(val.type()) {
    case value::value_type::VT_ERR:
    {
	std::ostringstream os;
	os  << "Failed to read " << print_list(path);
	throw std::runtime_error(os.str());
    }
    case value::value_type::VT_NULL:
	if(std::get<2>(place_)) {
	    std::ostringstream os;
	    os << "Required value for " << print_list(path) << " missing";
	    throw std::runtime_error(os.str());
	}
	// Use the default value
	check_type(std::get<1>(place_), path);
	return;
    default:
	// All remaining types
	check_type(val, path);
    }
}


void
variable::check_type(value const &v, list const &path) const
{
    char const *type = "unknown";
    if(debug)
	std::cerr << "CONF SET " << print_list(std::get<0>(place_)) << " TO ";
    switch(v.type()) {
    case value::value_type::VT_ERR:
    case value::value_type::VT_NULL:
    {
	std::ostringstream os;
	os << "Attempt to set" << print_list(path) << " to null/error";
	if(debug)
	    std::cerr << "NULL/ERR\n";
	throw std::runtime_error(os.str());
    }
    case value::value_type::VT_STR:
	if(s_) {
	    *s_ = v.get_str();
	    if(debug)
		std::cerr << *s_ << std::endl;
	    return;
	}
	type = "string";
	break;
    case value::value_type::VT_INT:
	if(i_) {
	    *i_ = v.get_int();
	    if(debug)
		std::cerr << *i_ << std::endl;
	    return;
	}
	type = "int";
	break;
    case value::value_type::VT_BOOL:
	if(b_) {
	    *b_ = v.get_bool();
	    if(debug)
		std::cerr << (*b_ ? "true" : "false") << std::endl;
	    return;
	}
	type = "bool";
	break;
    }
    // type mismatch
    std::ostringstream exc;
    exc << print_list(path) << ": expected ";
    if(s_) exc << "string";
    if(i_) exc << "int";
    if(b_) exc << "bool";
    exc << " got " << type;
    throw std::runtime_error(exc.str());
}

