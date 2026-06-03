import { Text, Tooltip } from "@mantine/core";

type Props = {
  value: string;
  maxWidth?: number;
};

export function TruncatedCell({ value, maxWidth = 220 }: Props) {
  const text = value.trim();
  const display = text || "—";

  return (
    <Tooltip
      label={display}
      multiline
      maw={480}
      disabled={!text}
      withArrow
      position="top-start"
    >
      <Text
        size="xs"
        truncate="end"
        maw={maxWidth}
        data-truncate={text ? "true" : "false"}
        data-testid="truncated-cell"
      >
        {display}
      </Text>
    </Tooltip>
  );
}
