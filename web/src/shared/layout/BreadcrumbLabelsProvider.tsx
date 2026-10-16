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
  setGone: (segment: string, gone: boolean) => void;
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
  const [gone, setGone] = React.useState<ReadonlySet<string>>(new Set());

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
    setGone: (segment, on) => {
      setGone((prev) => {
        if (prev.has(segment) === on) return prev;
        const next = new Set(prev);
        on ? next.add(segment) : next.delete(segment);
        return next;
      });
    },
  }).current;
  const shown = React.useMemo(
    () =>
      gone.size === 0
        ? labels
        : Object.fromEntries(Object.entries(labels).filter(([segment]) => !gone.has(segment))),
    [labels, gone],
  );

  return (
    <BreadcrumbLabelActionsContext.Provider value={actions}>
      <BreadcrumbLabelsValueContext.Provider value={shown}>
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

// While mounted, `segment`'s registered label is withheld: the record it named is gone, so the
// crumb reads the route's static label.
export function useBreadcrumbGone(segment: string | undefined) {
  const actions = React.useContext(BreadcrumbLabelActionsContext);

  React.useEffect(() => {
    if (!actions || !segment) return;
    actions.setGone(segment, true);
    return () => actions.setGone(segment, false);
  }, [actions, segment]);
}
