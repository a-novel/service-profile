import { decodeHttpResponse, handleHttpResponse } from "@a-novel-kit/nodelib-browser/http";

import type { ZodType } from "zod";

// Parses the JSON body as-is, trusting it to match T. Used when no schema is supplied.
async function decodeRawHttpResponse<T>(response: Response): Promise<T> {
  return await response.json();
}

/**
 * Health of a single upstream dependency reported by the service. The health endpoint is
 * unauthenticated, so the report carries the state alone. Diagnostics stay on the server's traces.
 */
export type HealthDependency = {
  status: "up" | "down";
};

/**
 * A ProfileApi is the HTTP client for the Profile service. Construct it with
 * the service base URL, then pass it to the resource helpers to make requests.
 */
export class ProfileApi {
  private readonly _baseUrl: string;

  constructor(baseUrl: string) {
    this._baseUrl = baseUrl;
  }

  /** Sends a request and resolves once the status is validated, discarding the body. */
  async fetchVoid(input: string, init?: RequestInit): Promise<void> {
    await fetch(`${this._baseUrl}${input}`, init).then(handleHttpResponse);
  }

  /** Sends a request and decodes the JSON body, validating it against the schema when given. */
  async fetch<T>(input: string, validator?: ZodType<T>, init?: RequestInit): Promise<T> {
    return await fetch(`${this._baseUrl}${input}`, init)
      .then(handleHttpResponse)
      .then(validator ? decodeHttpResponse(validator) : decodeRawHttpResponse<T>);
  }

  /** Sends a request and reads the body as a Blob, typed with the response's Content-Type. */
  async fetchBlob(input: string, init?: RequestInit): Promise<Blob> {
    return await fetch(`${this._baseUrl}${input}`, init)
      .then(handleHttpResponse)
      .then((response) => response.blob());
  }

  /** Liveness probe; resolves when the service is reachable. */
  async ping(): Promise<void> {
    await this.fetchVoid("/v1/ping", { method: "GET" });
  }

  /**
   * Reports the health of the service, keyed by upstream dependency name, when every dependency
   * is healthy. Throws on HTTP 503 when a dependency probe fails, or on another non-2xx response.
   */
  async health(): Promise<Record<string, HealthDependency>> {
    return await this.fetch("/v1/healthcheck", undefined, { method: "GET" });
  }
}
