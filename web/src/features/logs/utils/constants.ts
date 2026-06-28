export const LOG_LEVELS: Record<string, string> = {
  error: "bg-error/10 text-error ring-error/20",
  critical: "bg-error/10 text-error ring-error/20",
  info: "bg-info/10 text-info ring-info/30",
  warning: "bg-warning/10 text-warning ring-warning/20",
  warn: "bg-warning/10 text-warning ring-warning/20",
  success: "bg-success/10 text-success ring-success/20",
};

export const REFRESH_INTERVALS = [
  { label: "No refresh", value: "0" },
  { label: "15 seconds", value: "15000" },
  { label: "30 seconds", value: "30000" },
  { label: "1 minute", value: "60000" },
  { label: "5 minutes", value: "300000" },
] as const;

export const VALID_INTERVALS: string[] = REFRESH_INTERVALS.map((i) => i.value);

export enum TimeDuration {
  Minutes,
  Hours,
  Days,
}

export interface TimeOption {
  id: string;
  label: string;
  type: TimeDuration;
  duration: number;
}

export const TIME_OPTIONS: TimeOption[] = [
  {
    id: "5m",
    label: "5 mins",
    type: TimeDuration.Minutes,
    duration: 5,
  },
  {
    id: "15m",
    label: "15 mins",
    type: TimeDuration.Minutes,
    duration: 15,
  },
  {
    id: "30m",
    label: "30 mins",
    type: TimeDuration.Minutes,
    duration: 30,
  },
  {
    id: "1h",
    label: "1 hr",
    type: TimeDuration.Hours,
    duration: 1,
  },
  {
    id: "3h",
    label: "3 hrs",
    type: TimeDuration.Hours,
    duration: 3,
  },
  {
    id: "6h",
    label: "6 hrs",
    type: TimeDuration.Hours,
    duration: 6,
  },
  {
    id: "1d",
    label: "1 day",
    type: TimeDuration.Days,
    duration: 1,
  },
  {
    id: "2d",
    label: "2 days",
    type: TimeDuration.Days,
    duration: 2,
  },
  {
    id: "7d",
    label: "7 days",
    type: TimeDuration.Days,
    duration: 7,
  },
  {
    id: "15d",
    label: "15 days",
    type: TimeDuration.Days,
    duration: 15,
  },
  {
    id: "30d",
    label: "30 days",
    type: TimeDuration.Days,
    duration: 30,
  },
  {
    id: "60d",
    label: "60 days",
    type: TimeDuration.Days,
    duration: 60,
  },
  {
    id: "90d",
    label: "90 days",
    type: TimeDuration.Days,
    duration: 90,
  },
];
