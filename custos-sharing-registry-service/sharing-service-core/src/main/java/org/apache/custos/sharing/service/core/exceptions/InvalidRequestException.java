package org.apache.custos.sharing.service.core.exceptions;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ResponseStatus;

@ResponseStatus(value= HttpStatus.BAD_REQUEST, reason = "Request cannot be completed because the server refused to fulfill the request.")
public class InvalidRequestException extends RuntimeException{

    public InvalidRequestException(String errorMessage) {
        super(errorMessage);
    }

    public InvalidRequestException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
}
