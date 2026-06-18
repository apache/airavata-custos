"use client";

import { Line, LineChart, ResponsiveContainer } from "recharts";

export type SparklineProps = {
  data: Array<{ value: number }>;
  color?: string;
  height?: number;
  ariaLabel?: string;
};

export function Sparkline({
  data,
  color = "var(--custos-blue-500)",
  height = 32,
  ariaLabel,
}: SparklineProps) {
  return (
    <div role="img" aria-label={ariaLabel ?? "Sparkline"} className="w-full">
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 2, right: 2, left: 2, bottom: 2 }}>
          <Line
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={1.5}
            dot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
