package org.apache.custos.sharing.service.core.exceptions;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ResponseStatus;

@ResponseStatus(value= HttpStatus.NOT_FOUND, reason = "Request cannot be completed because the server could not find the resource")
public class ResourceNotFoundException extends RuntimeException{

    public ResourceNotFoundException(String errorMessage) {
        super(errorMessage);
    }

    public ResourceNotFoundException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
}
