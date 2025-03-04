package org.apache.custos.core.constants;

public enum Operation {
    CREATE("CREATE"),
    DELETE("DELETE"),
    UPDATE("UPDATE");

    private final String text;

    Operation(final String text) {
        this.text = text;
    }

    @Override
    public String toString() {
        return text;
    }
}
