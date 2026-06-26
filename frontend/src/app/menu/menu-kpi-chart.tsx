"use client";

import { useEffect, useRef, useState } from "react";
import { Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import type { ChartDatum } from "./menu-dashboard-data";

export function MenuKpiChart({ data, dictionary }: { data: ChartDatum[]; dictionary: BackendLanguageDictionary }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const [chartWidth, setChartWidth] = useState(0);

  useEffect(() => {
    const element = chartRef.current;
    if (!element) return;

    const updateWidth = () => setChartWidth(Math.max(0, Math.floor(element.getBoundingClientRect().width)));
    updateWidth();

    const observer = new ResizeObserver(updateWidth);
    observer.observe(element);

    return () => observer.disconnect();
  }, []);

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>{backendText(dictionary, "erp_menu_ratio")}</CardTitle>
        <CardDescription>{backendText(dictionary, "route_count_by_main_category")}</CardDescription>
      </CardHeader>
      <CardContent>
        <div ref={chartRef} className="h-64 min-h-64 min-w-0 w-full">
          {chartWidth > 0 ? (
            <BarChart data={data} height={256} margin={{ top: 8, right: 8, left: -24, bottom: 0 }} width={chartWidth}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
              <XAxis dataKey="name" tickLine={false} axisLine={false} fontSize={12} />
              <YAxis tickLine={false} axisLine={false} fontSize={12} />
              <Tooltip
                cursor={{ fill: "rgba(148, 163, 184, 0.12)" }}
                contentStyle={{
                  borderRadius: 8,
                  border: "1px solid var(--border)",
                  background: "var(--popover)",
                  color: "var(--popover-foreground)",
                }}
              />
              <Bar dataKey="value" fill="var(--primary)" radius={[10, 10, 4, 4]} />
            </BarChart>
          ) : (
            <div aria-hidden="true" className="h-full w-full animate-pulse rounded-lg bg-muted" />
          )}
        </div>
      </CardContent>
    </Card>
  );
}
