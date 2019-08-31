package org.apache.custos.sharing.service.core.exceptions;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ResponseStatus;

@ResponseStatus(value= HttpStatus.CONFLICT, reason = "Request cannot be completed because a duplicate entry is requested")
public class DuplicateEntryException extends RuntimeException {

    public DuplicateEntryException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
    public DuplicateEntryException(String errorMessage) {
        super(errorMessage);
    }
}
