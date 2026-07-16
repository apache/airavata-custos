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
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

const DEFAULT_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

export type StackedAreaUsageProps = {
  data: Array<{ date: string } & Record<string, string | number>>;
  seriesKeys: string[];
  colors?: string[];
  height?: number;
  ariaLabel?: string;
  // We extract (seriesKey, date) from Recharts' datum so callers don't depend
  // on Recharts payload internals.
  onSegmentClick?: (seriesKey: string, date: string) => void;
};

export function StackedAreaUsage({
  data,
  seriesKeys,
  colors = DEFAULT_COLORS,
  height = 240,
  ariaLabel,
  onSegmentClick,
}: StackedAreaUsageProps) {
  return (
    <div role="img" aria-label={ariaLabel ?? "Stacked area usage chart"}>
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 8 }}>
          <CartesianGrid stroke="var(--chart-grid)" />
          <XAxis dataKey="date" tick={{ fontSize: 12 }} stroke="var(--chart-grid)" />
          <YAxis tick={{ fontSize: 12 }} stroke="var(--chart-grid)" />
          <Tooltip />
          {seriesKeys.map((key, i) => (
            <Area
              key={key}
              type="monotone"
              dataKey={key}
              stackId="1"
              stroke={colors[i % colors.length]}
              fill={colors[i % colors.length]}
              fillOpacity={0.4}
              onClick={
                onSegmentClick
                  ? (props) => {
                      const payload = (props as { payload?: { date?: string } } | undefined)
                        ?.payload;
                      const date = payload?.date;
                      if (typeof date === "string") onSegmentClick(key, date);
                    }
                  : undefined
              }
              style={onSegmentClick ? { cursor: "pointer" } : undefined}
            />
          ))}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
