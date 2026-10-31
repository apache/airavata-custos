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

import Markdown, { type Components } from "react-markdown";

const components: Components = {
  h1: ({ children }) => <h2 className="mt-6 mb-2 text-lg font-semibold first:mt-0">{children}</h2>,
  h2: ({ children }) => (
    <h3 className="mt-5 mb-2 text-base font-semibold first:mt-0">{children}</h3>
  ),
  h3: ({ children }) => <h4 className="mt-4 mb-1.5 font-semibold first:mt-0">{children}</h4>,
  p: ({ children }) => <p className="my-2.5">{children}</p>,
  ol: ({ children }) => <ol className="my-2.5 list-decimal space-y-2 pl-6">{children}</ol>,
  ul: ({ children }) => <ul className="my-2.5 list-disc space-y-1.5 pl-6">{children}</ul>,
  a: ({ href, children }) => (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="text-[color:var(--brand)] underline underline-offset-2"
    >
      {children}
    </a>
  ),
};

// Headings start at h2 so the agreement title stays the page's only h1.
export function AgreementBody({ markdown }: { markdown: string }) {
  return (
    <div className="text-sm leading-relaxed text-foreground">
      <Markdown components={components}>{markdown}</Markdown>
    </div>
  );
}
