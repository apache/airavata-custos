package org.apache.custos.server.start;

import org.apache.custos.authentication.server.CustosAuthenticationServer;
import org.apache.custos.profile.server.ProfileServiceServer;

public class CustosAPIServerStarter {
    public static void main(String args[]) throws Exception {
        CustosAuthenticationServer custosAuthenticationServer = new CustosAuthenticationServer();
        ProfileServiceServer custosProfileServer = new ProfileServiceServer();

        custosAuthenticationServer.start();
        custosProfileServer.start();
    }
}
