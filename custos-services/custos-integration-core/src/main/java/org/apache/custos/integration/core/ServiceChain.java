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

package org.apache.custos.integration.core;

import java.util.ArrayList;
import java.util.List;


/**
 * A class represents the set of services that  needs to be invoked to
 * complete a full operation
 */
public final class ServiceChain {


private final List<ServiceTask> serviceTasks;


   private ServiceChain(ServiceChainBuilder serviceChainBuilder) {
       this.serviceTasks = serviceChainBuilder.serviceTasks;
   }

   public void serve(Object data){
       if (!serviceTasks.isEmpty()) {
           serviceTasks.get(0).invokeService(data);
       }
   }

   public static class ServiceChainBuilder {

       private ServiceCallback serviceCallback;
       private List<ServiceTask> serviceTasks = new ArrayList();
       private ServiceTask latestTask;

       private ServiceChainBuilder(ServiceTask firstTask, ServiceCallback serviceCallback) {
          this.latestTask = firstTask;
          this.serviceCallback = serviceCallback;
          this.latestTask.setServiceCallback(serviceCallback);
          this.serviceTasks.add(this.latestTask);
       }



       public ServiceChainBuilder nextTask(ServiceTask serviceTask){
           this.latestTask.setNextTask(serviceTask);
           this.latestTask = serviceTask;
           this.latestTask.setServiceCallback(serviceCallback);
           this.serviceTasks.add(this.latestTask);
           return this;

       }

       public ServiceChain build() {
           ServiceChain serviceChain = new ServiceChain(this);
           return serviceChain;
       }


    }

    public static ServiceChainBuilder newBuilder(ServiceTask firstTask, ServiceCallback serviceCallback) {
        return  new ServiceChainBuilder(firstTask, serviceCallback);
    }



}
