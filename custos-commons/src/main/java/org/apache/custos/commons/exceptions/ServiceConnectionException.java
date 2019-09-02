package org.apache.custos.commons.exceptions;

public class ServiceConnectionException extends RuntimeException {

    public ServiceConnectionException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
    public ServiceConnectionException(String errorMessage) {
        super(errorMessage);
    }
}
