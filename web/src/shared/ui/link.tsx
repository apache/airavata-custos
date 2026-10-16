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

import { focusHeadingOf } from "@/shared/ui/focus";
import NextLink from "next/link";
import type * as React from "react";

// Every in-app link: the destination's h1 takes focus once it renders, since the link itself
// (a crumb, a row, a notice's CTA) usually unmounts with the page it sat on.
export function Link({
  onClick,
  ...props
}: Omit<React.ComponentProps<typeof NextLink>, "href"> & { href: string }) {
  return (
    <NextLink
      {...props}
      onClick={(e) => {
        onClick?.(e);
        if (!e.defaultPrevented) focusHeadingOf(props.href);
      }}
    />
  );
}
