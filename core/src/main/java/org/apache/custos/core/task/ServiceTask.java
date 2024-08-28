/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.core.task;

/**
 * The ServiceTask interface represents a task that can be invoked and chained
 * with other tasks. It provides methods for invoking the service, invoking the
 * next task in the chain, setting the next task in the chain, and setting a
 * service callback.
 *
 * @param <T> the type of data input to the service
 * @param <U> the type of data output from the service
 */
public interface ServiceTask<T, U> {

    void invokeService(T data);

    void  invokeNextTask(U data);

    void setNextTask(ServiceTask<T, U> serviceTask);

    void setServiceCallback(ServiceCallback serviceCallback);

}
