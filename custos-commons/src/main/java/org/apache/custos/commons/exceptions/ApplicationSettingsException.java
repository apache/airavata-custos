package org.apache.custos.commons.exceptions;

public class ApplicationSettingsException extends CustosException {

	private static final long serialVersionUID = -4901850535475160411L;

	public ApplicationSettingsException(String message) {
		super(message);
	}
	
	public ApplicationSettingsException(String message, Throwable e) {
		super(message, e);
	}
}
