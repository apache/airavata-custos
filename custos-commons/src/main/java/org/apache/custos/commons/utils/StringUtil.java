/**
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.commons.utils;

import java.util.*;

public class StringUtil {
	
	public static Map<Integer, String> getContainedParameters(String s) {
		Map<Integer,String> parameterMap=new HashMap<Integer,String>();
		int i=0;
		for(i=0;i<s.length();i++){
			if (s.charAt(i)=='$' && (i+1)<s.length() && s.charAt(i+1)=='{'){
				int i2=s.indexOf('{', i+2);
				int e=s.indexOf('}', i+2);
				if (e!=-1){
					if (i2==-1 || e<i2){
						parameterMap.put(i, s.substring(i,e+1));
						i=e;
					}
				}
			}
		}
		return parameterMap;
	}
}