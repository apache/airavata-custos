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

"use client";

import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import { Toaster as Sonner, type ToasterProps, toast } from "sonner";

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      icons={{
        success: <CircleCheckIcon className="size-4" />,
        info: <InfoIcon className="size-4" />,
        warning: <TriangleAlertIcon className="size-4" />,
        error: <OctagonXIcon className="size-4" />,
        loading: <Loader2Icon className="size-4 animate-spin" />,
      }}
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast: "cn-toast",
        },
      }}
      {...props}
    />
  );
};

// Toasts that wait for a choice sit mid-screen; plain notifications stay in the default toaster.
const CONFIRM_TOASTER = "confirm";

const ConfirmToaster = () => (
  <Toaster id={CONFIRM_TOASTER} position="top-center" offset={{ top: "45vh" }} />
);

// An in-app stand-in for window.confirm: onConfirm runs only if the action is clicked.
function confirmToast(message: string, actionLabel: string, onConfirm: () => void) {
  toast.warning(message, {
    toasterId: CONFIRM_TOASTER,
    duration: Number.POSITIVE_INFINITY,
    action: { label: actionLabel, onClick: onConfirm },
    cancel: { label: "Cancel", onClick: () => {} },
  });
}

const toastError = (err: Error) => toast.error(err.message);

// Mutation callbacks for the success half; failures already toast from the query client.
const toastOnSuccess = (message: string, then?: () => void) => ({
  onSuccess: () => {
    toast.success(message);
    then?.();
  },
});

export { ConfirmToaster, Toaster, confirmToast, toastError, toastOnSuccess };
