package org.apache.custos.core.constants;

public enum Operations {
    CREATE("CREATE"),
    DELETE("DELETE"),
    UPDATE("UPDATE"),

    private final String text;

    Operations(final String text) {
        this.text = text;
    }


}
