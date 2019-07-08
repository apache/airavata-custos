namespace java org.apache.custos.commons.model.security
namespace php Custos.Commons.Model.Security
namespace cpp apache.custos.commons.model.security
namespace py custos.commons.model.security

/*
 * This file describes the definitions of the security model which encapsulates the information that needs to be passed
  to the API methods in order to authenticate and authorize the users.
 *
*/

struct AuthzToken {
    1: required string accessToken,
    2: optional map<string, string> claimsMap
}