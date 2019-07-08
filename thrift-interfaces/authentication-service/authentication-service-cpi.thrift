namespace java org.apache.custos.authentication.cpi
namespace php Custos.Authentication.CPI
namespace py custos.authentication.cpi

include "../custos-apis/security_model.thrift"
include "authentication_service_cpi_errors.thrift"
include "../custos-apis/userInfo_model.thrift"

service CustosAuthenticationService {
    bool isUserAuthenticated(1: required security_model.AuthzToken authzToken)
                            throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
    userInfo_model.UserInfo getUserInfoFromAuthzToken(1: required security_model.AuthzToken authzToken) throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
    security_model.AuthzToken getUserManagementServiceAccountAuthzToken(1: required security_model.AuthzToken authzToken, 2: required string gatewayId) throws (1: authentication_service_cpi_errors.CustosAuthenticationServiceException ae)
}
