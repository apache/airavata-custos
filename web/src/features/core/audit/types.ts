/**
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

import type { z } from "zod";
import type {
  deliveryStatusSchema,
  stepSchema,
  traceDeliverySchema,
  traceDetailSchema,
  traceListSchema,
  traceStatusSchema,
  traceSummarySchema,
} from "./schemas";

export type { StepNode } from "./schemas";

export type TraceStatus = z.infer<typeof traceStatusSchema>;
export type DeliveryStatus = z.infer<typeof deliveryStatusSchema>;
export type TraceSummary = z.infer<typeof traceSummarySchema>;
export type TraceListResponse = z.infer<typeof traceListSchema>;
export type Step = z.infer<typeof stepSchema>;
export type TraceDelivery = z.infer<typeof traceDeliverySchema>;
export type TraceDetail = z.infer<typeof traceDetailSchema>;

// What the list says about a trace, in the admin's words.
export type ListStatus = "failed" | "retrying" | "waiting" | "done";
