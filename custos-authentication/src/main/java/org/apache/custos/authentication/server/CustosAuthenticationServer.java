package org.apache.custos.authentication.server;

import org.apache.custos.authentication.cpi.CustosAuthenticationService;
import org.apache.custos.authentication.handler.CustosAuthenticationHandler;
import org.apache.custos.commons.utils.IServer;
import org.apache.custos.commons.utils.ServerSettings;
import org.apache.thrift.server.TServer;
import org.apache.thrift.server.TThreadPoolServer;
import org.apache.thrift.transport.TServerSocket;
import org.apache.thrift.transport.TServerTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;
import java.util.Date;;

public class CustosAuthenticationServer implements  IServer{

    private final static Logger logger = LoggerFactory.getLogger(CustosAuthenticationService.class);

    private static final String SERVER_NAME = "Authentication Service Server";
    private static final String SERVER_VERSION = "1.0";

    private ServerStatus status;
    private TServer server;

    public CustosAuthenticationServer() {
        setStatus(ServerStatus.STOPPED);
    }

    public void updateTime() {

    }

    public Date getTime() {
        return null;
    }

    public String getName() {
        return SERVER_NAME;
    }

    public String getVersion() {
        return SERVER_VERSION;
    }

    public void start() throws Exception {
        try {
            final int serverPort = Integer.parseInt(ServerSettings.getAuthenticationServerPort());
            final String serverHost = ServerSettings.getAuthenticationServerHost();
            CustosAuthenticationService.Processor authenticationProcessor = new CustosAuthenticationService.Processor(new CustosAuthenticationHandler());
            TServerTransport serverTransport;

            if (serverHost == null) {
                serverTransport = new TServerSocket(serverPort);
            } else {
                InetSocketAddress inetSocketAddress = new InetSocketAddress(serverHost, serverPort);
                serverTransport = new TServerSocket(inetSocketAddress);
            }
            TThreadPoolServer.Args options = new TThreadPoolServer.Args(serverTransport);
            options.minWorkerThreads = 30;
            server = new TThreadPoolServer(options.processor(authenticationProcessor));

            new Thread() {
                public void run() {
                    server.serve();
                    setStatus(ServerStatus.STOPPED);
                    logger.info("Authentication Service Server Stopped.");
                }
            }.start();
            new Thread() {
                public void run() {
                    while (!server.isServing()) {
                        try {
                            Thread.sleep(500);
                        } catch (InterruptedException e) {
                            break;
                        }
                    }
                    if (server.isServing()) {
                        setStatus(ServerStatus.STARTED);
                        logger.info("Starting Authentication Service Server on Port " + serverPort);
                        logger.info("Listening to Authentication Service Server clients ....");
                    }
                }
            }.start();
        }
         catch (TTransportException e) {
            setStatus(ServerStatus.FAILED);
            throw new Exception("Error while starting the Authentication Service Server", e);
        }
    }

    public void stop() throws Exception {

        if (server != null && server.isServing()) {
            setStatus(ServerStatus.STOPING);
            server.stop();
        }
    }

    public void restart() throws Exception {

        stop();
        start();
    }

    public void configure() throws Exception {

    }

    public ServerStatus getStatus() throws Exception {
        return status;
    }

    private void setStatus(ServerStatus stat) {
        status = stat;
        status.updateTime();
    }

    public TServer getServer() {
        return server;
    }

    public void setServer(TServer server) {
        this.server = server;
    }

    public static void main(String[] args) {
        try {
            new CustosAuthenticationServer().start();
        } catch (Exception e) {
            logger.error(e.getMessage(), e);
        }
    }
}
