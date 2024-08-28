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
 * ServiceTaskImpl is an abstract class that implements the ServiceTask interface.
 * This implements invokeNextTask logic
 *
 * @param <T> the type of data input to the service
 * @param <U> the type of data output from the service
 */
public abstract class ServiceTaskImpl<T, U> implements ServiceTask<T, U> {

    private ServiceCallback serviceCallback;

    private ServiceTask<T, U> nextTask;

    private final String SUCCESS_TASK_STATUS = "Success";


    public void setServiceCallback(ServiceCallback serviceCallback) {
        this.serviceCallback = serviceCallback;
    }

    public ServiceCallback getServiceCallback() {
        return serviceCallback;
    }

    public void setNextTask(ServiceTask<T, U> nextTask) {
        this.nextTask = nextTask;
    }

    public ServiceTask<T, U> getNextTask() {
        return nextTask;
    }

    @Override
    public void invokeNextTask(U output) {
        if (nextTask != null) {
            nextTask.invokeService((T) output);
        } else {
            System.out.println("Calling Parent Callback");
            serviceCallback.onCompleted(SUCCESS_TASK_STATUS);
        }
    }
}
