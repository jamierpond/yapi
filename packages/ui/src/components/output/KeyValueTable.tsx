"use client";

export interface KeyValueRow {
  key: string;
  // Using any to avoid React 18/19 ReactNode type conflicts across packages
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
}

interface KeyValueTableProps {
  headers: [string, string];
  rows: KeyValueRow[];
}

export default function KeyValueTable({ headers, rows }: KeyValueTableProps) {
  return (
    <div className="h-full overflow-auto">
      <table className="w-full">
        <thead className="sticky top-0 bg-yapi-bg-elevated/80 backdrop-blur-sm border-b border-yapi-border/50">
          <tr>
            <th className="text-left text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider px-6 py-3">
              {headers[0]}
            </th>
            <th className="text-left text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider px-6 py-3">
              {headers[1]}
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row, idx) => (
            <tr
              key={row.key}
              className={`border-b border-yapi-border/20 hover:bg-yapi-bg-elevated/30 transition-colors ${
                idx % 2 === 0 ? "bg-yapi-bg/50" : "bg-yapi-bg-elevated/10"
              }`}
            >
              <td className="px-6 py-3 text-xs font-mono font-semibold text-yapi-accent">
                {row.key}
              </td>
              <td className="px-6 py-3 text-xs font-mono text-yapi-fg break-all">
                {row.value}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
