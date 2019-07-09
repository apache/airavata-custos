import org.apache.custos.authentication.server.CustosAuthenticationServer;

public class CustosAPIServerStarted {
    public static void main(String args[]) throws Exception {
        CustosAuthenticationServer custosAuthenticationServer = new CustosAuthenticationServer();
        custosAuthenticationServer.start();
    }
}
