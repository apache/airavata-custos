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
import { z } from "zod";

const DEFAULT_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

// Recharts hands the click handler the clicked datum under `payload`, which its types leave out.
const clickedDatumSchema = z.object({ payload: z.object({ date: z.string() }) });

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
          {seriesKeys.map((key, i) => {
            const color = colors[i % colors.length];
            return (
              <Area
                key={key}
                type="monotone"
                dataKey={key}
                stackId="1"
                {...(color === undefined ? {} : { stroke: color, fill: color })}
                fillOpacity={0.4}
                {...(onSegmentClick
                  ? {
                      onClick: (props: unknown) => {
                        const clicked = clickedDatumSchema.safeParse(props);
                        if (clicked.success) onSegmentClick(key, clicked.data.payload.date);
                      },
                      style: { cursor: "pointer" },
                    }
                  : {})}
              />
            );
          })}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
