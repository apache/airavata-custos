package org.apache.custos.client.authentication.service;

import org.apache.custos.authentication.cpi.CustosAuthenticationService;
import org.apache.custos.authentication.cpi.exception.CustosAuthenticationServiceException;
import org.apache.custos.commons.model.error.AuthenticationException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TMultiplexedProtocol;
import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.transport.TSocket;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

public class AuthenticationServiceClient {
    public static CustosAuthenticationService.Client createAuthenticationServiceClient(String serverHost, int serverPort)  throws CustosAuthenticationServiceException {
        try {
            TTransport transport = new TSocket(serverHost, serverPort);
            transport.open();
            TProtocol protocol = new TBinaryProtocol(transport);
            TMultiplexedProtocol multiplexedProtocol = new TMultiplexedProtocol(protocol, "CustosAuthenticationService");
            return new CustosAuthenticationService.Client(multiplexedProtocol);
        } catch (TTransportException e) {
            throw new CustosAuthenticationServiceException(e.getMessage());
        }
    }
}
