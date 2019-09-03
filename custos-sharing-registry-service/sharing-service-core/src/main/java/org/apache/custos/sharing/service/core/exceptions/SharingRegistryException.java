package org.apache.custos.sharing.service.core.exceptions;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ResponseStatus;

@ResponseStatus(value= HttpStatus.INTERNAL_SERVER_ERROR, reason = "Request cannot be completed due to server error")
public class SharingRegistryException extends RuntimeException {

    public SharingRegistryException(String errorMessage) {
        super(errorMessage);
    }

    public SharingRegistryException(String errorMessage, Throwable err) {
        super(errorMessage, err);
    }
}
