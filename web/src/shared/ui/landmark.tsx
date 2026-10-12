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

// Popups (selects, menus, popovers) portal into the landmark or open dialog holding their trigger,
// so every popup stays inside a page region; with no such ancestor they fall back to body. Callers
// position them fixed so the landmark's own scrolling cannot clip them.

"use client";

import * as React from "react";

const LANDMARKS = '[role="dialog"],[role="alertdialog"],main,header,nav,aside,footer';

type Landmark = {
  anchor: (trigger: HTMLElement | null) => void;
  container: React.RefObject<HTMLElement | null>;
  trigger: React.RefObject<HTMLElement | null>;
};

export function useLandmark(): Landmark {
  const container = React.useRef<HTMLElement | null>(null);
  const trigger = React.useRef<HTMLElement | null>(null);
  return React.useMemo(
    () => ({
      anchor: (el) => {
        if (!el) return;
        trigger.current = el;
        container.current = el.closest<HTMLElement>(LANDMARKS);
      },
      container,
      trigger,
    }),
    [],
  );
}

const LandmarkContext = React.createContext<Landmark | undefined>(undefined);

// For primitives whose trigger and popup are separate components under one root.
export function LandmarkScope({ children }: { children: React.ReactNode }) {
  return <LandmarkContext.Provider value={useLandmark()}>{children}</LandmarkContext.Provider>;
}

export const useLandmarkScope = () => React.useContext(LandmarkContext);
