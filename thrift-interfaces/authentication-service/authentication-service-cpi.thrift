namespace java org.apache.custos.authentication.cpi
namespace php Custos.Authentication.CPI
namespace py custos.authentication.cpi

include "security_model.thrift"
include "authentication_service_cpi_errors.thrift"
include "userInfo_model.thrift"

service CustosAuthenticationService {
    bool isUserAuthenticated(1: required security_model.AuthzToken authzToken)
                            throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
    userInfo_model.UserInfo getUserInfoFromAuthzToken(1: required security_model.AuthzToken authzToken) throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
    security_model.AuthzToken getUserManagementServiceAccountAuthzToken(1: required security_model.AuthzToken authzToken, 2: required string gatewayId, 3:required string clientId, 4: required string clientSecret ) throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
}
