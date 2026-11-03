import { describe, expect, it } from "vitest";

import { ProfileApi } from "@a-novel/service-profile-rest";

const api = new ProfileApi(process.env.REST_URL!);

describe("ping", () => {
  it("returns success", async () => {
    await expect(api.ping()).resolves.toBeUndefined();
  });
});

describe("health", () => {
  it("returns success", async () => {
    await expect(api.health()).resolves.toEqual({
      "client:postgres": { status: "up" },
    });
  });
});
