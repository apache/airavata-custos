package org.apache.custos.core.services.commons.util;

public class MethodNameExtractor {

    public static String getMethodName (String methodName) {
        String[] names = methodName.split("/");
        return  names[names.length - 1];
    }
}
