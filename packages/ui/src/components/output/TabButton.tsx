"use client";

interface TabButtonProps {
  label: string;
  isActive: boolean;
  onClick: () => void;
  count?: number;
  countVariant?: "default" | "warning";
}

export default function TabButton({
  label,
  isActive,
  onClick,
  count,
  countVariant = "default",
}: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 text-xs font-medium rounded-t-lg transition-all ${
        isActive
          ? "bg-yapi-bg text-yapi-accent border-b-2 border-yapi-accent"
          : "text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated/50"
      }`}
    >
      {label}
      {count !== undefined && (
        <span
          className={`ml-2 px-1.5 py-0.5 text-[10px] rounded ${
            countVariant === "warning"
              ? "bg-yapi-warning/20 text-yapi-warning"
              : "bg-yapi-bg-elevated/70"
          }`}
        >
          {count}
        </span>
      )}
    </button>
  );
}
