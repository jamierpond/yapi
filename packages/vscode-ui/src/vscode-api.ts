/**
 * VS Code API
 *
 * This module provides a typed wrapper around the VS Code webview API.
 * The API is acquired once and cached.
 */

type VSCodeAPI = ReturnType<typeof acquireVsCodeApi>;

let vscodeApi: VSCodeAPI | undefined;

export function getVSCodeAPI(): VSCodeAPI {
  if (!vscodeApi) {
    vscodeApi = acquireVsCodeApi();
  }
  return vscodeApi;
}

// Message types from webview to extension
export type WebviewMessage =
  | { type: 'ready' }
  | { type: 'requestData' };

// Message types from extension to webview
export type ExtensionMessage =
  | { type: 'setLoading'; loading: boolean }
  | { type: 'setResult'; result: any }
  | { type: 'clearResult' };
