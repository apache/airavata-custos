namespace java org.apache.custos.security.model
namespace php Custos.Security.Model
namespace cpp apache.airavata.security.model
namespace py airavata.security.model

/*
 * This file describes the definitions of the security model which encapsulates the information that needs to be passed
  to the API methods in order to authenticate and authorize the users.
 *
*/

struct AuthzToken {
    1: required string accessToken,
    2: optional map<string, string> claimsMap
}