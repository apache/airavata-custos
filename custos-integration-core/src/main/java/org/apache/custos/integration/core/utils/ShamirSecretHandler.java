/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.integration.core.utils;

import com.codahale.shamir.Scheme;
import com.google.protobuf.ByteString;
import org.apache.custos.integration.core.exceptions.InValidParameterException;

import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

public class ShamirSecretHandler {

    public static Map<Integer, byte[]> splitSecret(String secret, int numOfsplits, int threshold) {

        if (numOfsplits >= threshold) {
            Scheme scheme = new Scheme(new SecureRandom(), numOfsplits, threshold);
            byte[] secretArray = secret.getBytes(StandardCharsets.UTF_8);
            final Map<Integer, byte[]> parts =  scheme.split(secretArray);
            final Map<Integer, byte[]> selectedParts = new HashMap<>();
            int count  = 1;
            for (Integer integer : parts.keySet()){
                selectedParts.put(integer, parts.get(integer));
                count ++;
                if (count == threshold){
                    break;
                }
            }
            return  selectedParts;
        } else {
            throw new
                    InValidParameterException(
                    "Cannot split message number of splits should be greater than threshold", null);
        }
    }


    public static  String generateSecret(List<ByteString> byteStringList, int numOfsplits, int threshold) {
        Scheme scheme = new Scheme(new SecureRandom(), numOfsplits, threshold);
        Map<Integer, byte[]> selectedSplits = new HashMap<>();
        AtomicInteger count = new AtomicInteger();
        byteStringList.forEach(str-> {
            selectedSplits.put(count.get(), str.toByteArray());
            count.getAndIncrement();
        });

        final byte[] recovered = scheme.join(selectedSplits);
        return new String(recovered, StandardCharsets.UTF_8);
    }

}
