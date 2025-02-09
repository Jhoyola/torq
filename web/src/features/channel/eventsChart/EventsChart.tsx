// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "features/charts/useD3";
import { useEffect, useState } from "react";
import { Selection } from "d3";
import { ChartCanvas, EventsPlot, LinePlot, BarPlot } from "features/charts/charts";
import "features/charts/chart.scss";
import { useGetSettingsQuery } from "apiSlice";
import { ChannelEventResponse } from "features/channel/channelTypes";

export const EventChartKeyOptions = [
  { value: "amount", label: "Amount" },
  { value: "revenue", label: "Revenue" },
  { value: "count", label: "Count" },
];

type EventsChart = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any[];
  eventKey: string;
  events: ChannelEventResponse;
  selectedEventTypes: {
    feeRate: boolean;
    baseFee: boolean;
    minHtlc: boolean;
    maxHtlc: boolean;
    enabled: boolean;
    disabled: boolean;
  };
  from: string;
  to: string;
};

function EventsChart({ data, eventKey, events, selectedEventTypes, from, to }: EventsChart) {
  let chart: ChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];
  const [eventKeyLabel, setEventKeyLabel] = useState<string>("Amount");
  const settings = useGetSettingsQuery();

  // Check and update the chart size if the navigation changes the container size
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const navCheck = (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
    return () => {
      const boundingBox = container?.node()?.getBoundingClientRect();
      if (currentSize[0] !== boundingBox?.width || currentSize[1] !== boundingBox?.height) {
        chart.resizeChart();
        chart.draw();
        currentSize = [boundingBox?.width, boundingBox?.height];
      }
    };
  };

  useEffect(() => {
    setEventKeyLabel(EventChartKeyOptions.find((d) => d.value === eventKey)?.label || "Amount");
  }, [eventKey]);

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
      chart = new ChartCanvas(container, data, {
        from: new Date(from),
        to: new Date(to),
        timezone: settings?.data?.preferredTimezone || "UTC",
        yScaleKey: eventKey + "Total",
        rightYScaleKey: eventKey + "Total",
        rightYAxisKeys: [eventKey + "Out", eventKey + "In", eventKey + "Total"],
        xAxisPadding: 12,
      });
      chart.plot(BarPlot, {
        id: eventKey + "Total",
        key: eventKey + "Total",
        legendLabel: eventKeyLabel + " Total",
        barColor: "rgba(133, 196, 255, 0.5)",
        // areaGradient: ["rgba(133, 196, 255, 0.5)", "rgba(87, 211, 205, 0.5)"],
      });
      chart.plot(LinePlot, {
        id: eventKey + "Out",
        key: eventKey + "Out",
        legendLabel: eventKeyLabel + " Out",
        lineColor: "#BA93FA",
        // rightAxis: true,
      });
      chart.plot(LinePlot, {
        id: eventKey + "In",
        key: eventKey + "In",
        legendLabel: eventKeyLabel + " In",
        lineColor: "#FAAE93",
      });
      const filteredEvents =
        events?.events?.filter((d) => {
          switch (d.type) {
            case "fee_rate":
              return selectedEventTypes.feeRate;
            case "base_fee":
              return selectedEventTypes.baseFee;
            case "min_htlc":
              return selectedEventTypes.minHtlc;
            case "max_htlc":
              return selectedEventTypes.maxHtlc;
            case "rebalanced":
              return false;
            case "disabled":
              return selectedEventTypes.disabled;
            case "enabled":
              return selectedEventTypes.enabled;
          }
          return selectedEventTypes[d.type as keyof typeof selectedEventTypes]; // selectedEventTypes
        }) || [];
      chart.plot(EventsPlot, { id: "events", key: "events", events: filteredEvents });
      chart.draw();
      setInterval(navCheck(container), 200);
    },
    [
      data,
      eventKey,
      data ? data[0].date : "",
      data ? data[data.length - 1].date : "",
      selectedEventTypes,
      settings,
      eventKeyLabel,
    ]
  );

  useEffect(() => {
    return () => {
      if (chart) {
        chart.removeResizeListener();
      }
    };
  }, [data, data ? data[0].date : ""]);

  return <div ref={ref} className={"testing"} />;
}

export default EventsChart;
