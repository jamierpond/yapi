"use client";

import { useState, useCallback } from "react";
import type { ExecuteResponse } from "../types/api-contract";
import { isSuccessResponse } from "../types/api-contract";
import JsonViewer from "./JsonViewer";
import {
  LoadingSkeleton,
  EmptyState,
  ResponseHeader,
  TabButton,
  KeyValueTable,
  ErrorDisplay,
  WarningsDisplay,
  TestResultsTab,
  type KeyValueRow,
} from "./output";

interface OutputPanelProps {
  result: ExecuteResponse | null;
  isLoading: boolean;
}

type Tab = "body" | "headers" | "cookies" | "info" | "warnings" | "tests";

export default function OutputPanel({ result, isLoading }: OutputPanelProps) {
  const [activeTab, setActiveTab] = useState<Tab>("body");

  const getBodyForCopy = useCallback(() => {
    if (!result || !isSuccessResponse(result)) return undefined;
    return typeof result.responseBody === "string"
      ? result.responseBody
      : JSON.stringify(result.responseBody, null, 2);
  }, [result]);

  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!result) {
    return <EmptyState />;
  }

  const isSuccess = isSuccessResponse(result);

  return (
    <div className="h-full flex flex-col bg-yapi-bg relative">
      <div className="flex-1 flex flex-col overflow-hidden">
        <ResponseHeader
          statusCode={isSuccess ? result.statusCode : undefined}
          timing={isSuccess ? result.timing : 0}
          onCopy={isSuccess ? getBodyForCopy : undefined}
          assertions={isSuccess ? result.assertions : undefined}
        />

        {isSuccess ? (
          <SuccessContent
            result={result}
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
        ) : (
          <ErrorDisplay
            errorType={result.errorType}
            error={result.error}
            details={result.details}
          />
        )}
      </div>
    </div>
  );
}

interface SuccessContentProps {
  result: Extract<ExecuteResponse, { success: true }>;
  activeTab: Tab;
  setActiveTab: (tab: Tab) => void;
}

function SuccessContent({ result, activeTab, setActiveTab }: SuccessContentProps) {
  const hasHeaders = result.headers && Object.keys(result.headers).length > 0;
  const hasCookies = hasHeaders && Object.keys(result.headers!).some(
    (key) => key.toLowerCase() === "set-cookie"
  );
  const hasWarnings = result.warnings && result.warnings.length > 0;
  const hasTests = result.assertions && result.assertions.total > 0;
  const testsFailed = hasTests && result.assertions!.passed < result.assertions!.total;

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-yapi-bg">
      {/* Tabs */}
      <div className="flex items-center gap-1 px-4 pt-2 border-b border-yapi-border/30 bg-yapi-bg-elevated/30">
        <TabButton
          label="Body"
          isActive={activeTab === "body"}
          onClick={() => setActiveTab("body")}
        />
        {hasTests && (
          <TabButton
            label="Tests"
            isActive={activeTab === "tests"}
            onClick={() => setActiveTab("tests")}
            count={result.assertions!.total}
            countVariant={testsFailed ? "warning" : "default"}
          />
        )}
        {hasHeaders && (
          <>
            <TabButton
              label="Headers"
              isActive={activeTab === "headers"}
              onClick={() => setActiveTab("headers")}
              count={Object.keys(result.headers!).length}
            />
            {hasCookies && (
              <TabButton
                label="Cookies"
                isActive={activeTab === "cookies"}
                onClick={() => setActiveTab("cookies")}
              />
            )}
          </>
        )}
        <TabButton
          label="Info"
          isActive={activeTab === "info"}
          onClick={() => setActiveTab("info")}
        />
        {hasWarnings && (
          <TabButton
            label="Warnings"
            isActive={activeTab === "warnings"}
            onClick={() => setActiveTab("warnings")}
            count={result.warnings!.length}
            countVariant="warning"
          />
        )}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === "body" && (
          <JsonViewer
            value={
              typeof result.responseBody === "string"
                ? result.responseBody
                : JSON.stringify(result.responseBody, null, 2)
            }
          />
        )}

        {activeTab === "tests" && result.assertions && (
          <TestResultsTab assertions={result.assertions} />
        )}

        {activeTab === "headers" && result.headers && (
          <KeyValueTable
            headers={["Name", "Value"]}
            rows={Object.entries(result.headers).map(([key, value]) => ({
              key,
              value,
            }))}
          />
        )}

        {activeTab === "cookies" && result.headers && (
          <CookiesTable headers={result.headers} />
        )}

        {activeTab === "info" && <InfoTable result={result} />}

        {activeTab === "warnings" && result.warnings && (
          <WarningsDisplay warnings={result.warnings} />
        )}
      </div>
    </div>
  );
}

function CookiesTable({ headers }: { headers: Record<string, string> }) {
  const cookies = Object.entries(headers)
    .filter(([key]) => key.toLowerCase() === "set-cookie")
    .flatMap(([_, value]) => {
      const parts = value.split(";").map((p) => p.trim());
      const [nameValue, ...attributes] = parts;
      const [name, cookieValue] = nameValue.split("=");
      return [
        {
          name: name.trim(),
          value: cookieValue?.trim() || "",
          attributes: attributes.join("; "),
        },
      ];
    });

  return (
    <div className="h-full overflow-auto">
      <table className="w-full">
        <thead className="sticky top-0 bg-yapi-bg-elevated/80 backdrop-blur-sm border-b border-yapi-border/50">
          <tr>
            <th className="text-left text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider px-6 py-3">
              Name
            </th>
            <th className="text-left text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider px-6 py-3">
              Value
            </th>
            <th className="text-left text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider px-6 py-3">
              Attributes
            </th>
          </tr>
        </thead>
        <tbody>
          {cookies.map((cookie, idx) => (
            <tr
              key={`${cookie.name}-${idx}`}
              className={`border-b border-yapi-border/20 hover:bg-yapi-bg-elevated/30 transition-colors ${
                idx % 2 === 0 ? "bg-yapi-bg/50" : "bg-yapi-bg-elevated/10"
              }`}
            >
              <td className="px-6 py-3 text-xs font-mono font-semibold text-yapi-accent">
                {cookie.name}
              </td>
              <td className="px-6 py-3 text-xs font-mono text-yapi-fg break-all">
                {cookie.value}
              </td>
              <td className="px-6 py-3 text-xs font-mono text-yapi-fg-muted">
                {cookie.attributes || "-"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function InfoTable({ result }: { result: Extract<ExecuteResponse, { success: true }> }) {
  const rows = [
    result.transport && {
      key: "Transport",
      value: (
        <span className="px-2 py-1 text-xs font-mono font-semibold bg-yapi-accent/10 text-yapi-accent border border-yapi-accent/30 rounded">
          {result.transport.toUpperCase()}
        </span>
      ),
    },
    result.requestUrl && { key: "URL", value: result.requestUrl },
    result.method && {
      key: "Method",
      value: (
        <span className="px-2 py-1 text-xs font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/30 rounded">
          {result.method}
        </span>
      ),
    },
    result.service && { key: "Service", value: result.service },
    result.statusCode !== undefined && {
      key: "Status",
      value: (
        <span
          className={`px-2 py-1 text-xs font-mono font-semibold rounded ${
            result.statusCode >= 200 && result.statusCode < 300
              ? "bg-yapi-success/10 text-yapi-success border border-yapi-success/30"
              : result.statusCode >= 400
              ? "bg-yapi-error/10 text-yapi-error border border-yapi-error/30"
              : "bg-yapi-warning/10 text-yapi-warning border border-yapi-warning/30"
          }`}
        >
          {result.statusCode}
        </span>
      ),
    },
    { key: "Time", value: `${result.timing}ms` },
    result.sizeBytes !== undefined && {
      key: "Size",
      value: (
        <>
          {result.sizeBytes} bytes
          {result.sizeLines !== undefined && ` / ${result.sizeLines} lines`}
          {result.sizeChars !== undefined && ` / ${result.sizeChars} chars`}
        </>
      ),
    },
    result.contentType && { key: "Content-Type", value: result.contentType },
  ].filter(Boolean) as KeyValueRow[];

  return (
    <KeyValueTable
      headers={["Property", "Value"]}
      rows={rows}
    />
  );
}
