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

import * as React from "react";

type BreadcrumbLabelActions = {
  setLabel: (segment: string, label: string) => void;
  clearLabel: (segment: string) => void;
};

// Split into two contexts on purpose: the actions object is created once and
// never changes identity, so components that only register a label (via
// useBreadcrumbLabel) never re-render when someone else's label changes.
// Only useBreadcrumbLabels (read by Breadcrumbs) subscribes to the data
// itself. Merging these into one context would make the actions' identity
// change whenever the labels record changes, which — combined with an effect
// that both depends on and calls those actions — is a self-sustaining
// render loop.
const BreadcrumbLabelsValueContext = React.createContext<Record<string, string>>({});
const BreadcrumbLabelActionsContext = React.createContext<BreadcrumbLabelActions | null>(null);

export function BreadcrumbLabelsProvider({ children }: { children: React.ReactNode }) {
  const [labels, setLabels] = React.useState<Record<string, string>>({});

  const actions = React.useRef<BreadcrumbLabelActions>({
    setLabel: (segment, label) => {
      setLabels((prev) => (prev[segment] === label ? prev : { ...prev, [segment]: label }));
    },
    clearLabel: (segment) => {
      setLabels((prev) => {
        if (!(segment in prev)) return prev;
        const next = { ...prev };
        delete next[segment];
        return next;
      });
    },
  }).current;

  return (
    <BreadcrumbLabelActionsContext.Provider value={actions}>
      <BreadcrumbLabelsValueContext.Provider value={labels}>
        {children}
      </BreadcrumbLabelsValueContext.Provider>
    </BreadcrumbLabelActionsContext.Provider>
  );
}

export function useBreadcrumbLabels(): Record<string, string> {
  return React.useContext(BreadcrumbLabelsValueContext);
}

// Registers a human-readable breadcrumb label for a dynamic route segment
// (e.g. a resource's name in place of its raw ID) for as long as the calling
// component is mounted with that segment/label pair.
export function useBreadcrumbLabel(segment: string | undefined, label: string | undefined) {
  const actions = React.useContext(BreadcrumbLabelActionsContext);

  React.useEffect(() => {
    if (!actions || !segment || !label) return;
    actions.setLabel(segment, label);
    return () => actions.clearLabel(segment);
  }, [actions, segment, label]);
}
