package org.apache.custos.rest.controller;

import org.apache.custos.client.profile.service.CustosProfileServiceClientFactory;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.profile.model.workspace.Gateway;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.rest.resources.GatewayResource;
import org.apache.thrift.TException;
import org.dozer.DozerBeanMapper;
import org.jboss.resteasy.client.jaxrs.ResteasyClientBuilder;
import org.keycloak.admin.client.KeycloakBuilder;
import org.keycloak.representations.idm.RealmRepresentation;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.keycloak.admin.client.Keycloak;

@RestController
@RequestMapping("/tenant")
public class TenantController {

    @Autowired
    private TenantProfileService.Client tenantClient;

    @Autowired
    private DozerBeanMapper mapper;

    @Autowired
    private Keycloak keycloakAdminClient;

    private static AuthzToken authzToken = new AuthzToken("empy_token");

    /*

    Sample Request

    POST
    http://localhost:8080/tenant/

        {
          "gatewayId":"test-gateway-10",
          "gatewayApprovalStatus":"APPROVED",
          "gatewayName":"test-gateway-10",
          "domain":"test-gateway-domain",
          "emailAddress":"test-gateway-1@gmail.com",
          "gatewayURL":"test-gateway-1.com",
          "gatewayAdminFirstName":"John",
          "gatewayAdminLastName":"Doe",
          "gatewayAdminEmail":"admin.test-gateway-1@gmail.com"
        }
     */
    @RequestMapping(value = "", method = RequestMethod.POST)
    public String createTenant(@RequestBody GatewayResource gatewayResource) throws TException {

        Gateway gatewayT = mapper.map(gatewayResource, Gateway.class);
        Gateway gateway = tenantClient.addGateway(authzToken, gatewayT);
        return gateway.getCustosInternalGatewayId();
    }

    /*
    Sample Request

    GET
    Creating a realm on keycloak for above gateway
        http://localhost:8080/tenant/realm/40e5be38-0fde-41fe-a846-13de8b2ecfec
     */
    @RequestMapping(value = "/realm/{tenant}", method = RequestMethod.POST)
    public String createRealm(@PathVariable("tenant") String gateway) throws TException {
        Gateway gatewayT = tenantClient.getGateway(authzToken, gateway);

        RealmRepresentation realmRepresentation = new RealmRepresentation();
        realmRepresentation.setRealm(gatewayT.getGatewayId());
        keycloakAdminClient.realms().create(realmRepresentation);

        return "Created";
    }


}
