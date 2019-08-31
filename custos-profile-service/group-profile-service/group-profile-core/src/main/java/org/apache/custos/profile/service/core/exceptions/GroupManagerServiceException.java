package org.apache.custos.profile.service.core.exceptions;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ResponseStatus;

@ResponseStatus(value= HttpStatus.INTERNAL_SERVER_ERROR, reason = "Request cannot be completed due to server error")
public class GroupManagerServiceException extends RuntimeException {

    public GroupManagerServiceException(String errorMessage) {
        super(errorMessage);
    }

    public GroupManagerServiceException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
}
