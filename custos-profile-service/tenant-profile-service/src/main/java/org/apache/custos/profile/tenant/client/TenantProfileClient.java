package org.apache.custos.profile.tenant.client;

import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.profile.tenant.cpi.exception.TenantProfileServiceException;
import org.apache.custos.profile.tenant.cpi.profile_tenant_cpiConstants;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TMultiplexedProtocol;
import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.transport.TSocket;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

public class TenantProfileClient {

    public static TenantProfileService.Client createCustosTenantProfileServiceClient(String serverHost, int serverPort) throws TenantProfileServiceException {
        try {
            TTransport transport = new TSocket(serverHost, serverPort);
            transport.open();
            TProtocol protocol = new TBinaryProtocol(transport);
            TMultiplexedProtocol multiplexedProtocol = new TMultiplexedProtocol(protocol, profile_tenant_cpiConstants.TENANT_PROFILE_CPI_NAME);
            return new TenantProfileService.Client(multiplexedProtocol);
        } catch (TTransportException e) {
            throw new TenantProfileServiceException(e.getMessage());
        }
    }
}
