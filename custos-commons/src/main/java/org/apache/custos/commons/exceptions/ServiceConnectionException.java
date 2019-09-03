package org.apache.custos.commons.exceptions;
/*
This exception is thrown when client fails to make the Http call
*/

public class ServiceConnectionException extends RuntimeException {

    public ServiceConnectionException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
    public ServiceConnectionException(String errorMessage) {
        super(errorMessage);
    }
}
