// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { AlertCircleIcon } from "lucide-react";
import * as React from "react";

export type ErrorStateProps = {
  heading?: string;
  message?: React.ReactNode;
  onRetry?: () => void;
  retryLabel?: string;
  className?: string;
};

export function ErrorState({
  heading = "Something went wrong",
  message,
  onRetry,
  retryLabel = "Try again",
  className,
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className={cn(
        // Soft-destructive tint; text stays neutral so contrast on the tint hits AA.
        "flex flex-col items-center justify-center gap-3 rounded-md border border-[color:var(--custos-red-200)] bg-[color:var(--custos-red-50)] px-6 py-10 text-center",
        className,
      )}
    >
      <AlertCircleIcon className="size-8 text-[color:var(--custos-red-600)]" aria-hidden="true" />
      <h3 className="font-heading text-base font-medium text-foreground">{heading}</h3>
      {message ? <p className="max-w-md text-sm text-muted-foreground">{message}</p> : null}
      {onRetry ? (
        <Button variant="outline" onClick={onRetry}>
          {retryLabel}
        </Button>
      ) : null}
    </div>
  );
}
