// @vitest-environment node
import { anonymousAccessToken, userAccessToken } from "./token";

import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { isHttpError } from "@a-novel-kit/nodelib-browser/http";
import { expectStatus } from "@a-novel-kit/nodelib-test/http";
import {
  PictureRefusal,
  ProfileApi,
  profileGet,
  profileGetSelf,
  profilePictureDelete,
  profilePictureGet,
  profilePictureUpdate,
  profileUsernameUpdate,
} from "@a-novel/service-profile-rest";

const api = new ProfileApi(process.env.REST_URL!);

/** A 16 px square WebP, which the server accepts. */
const square = new Blob([readFileSync(new URL("../testdata/square.webp", import.meta.url))], { type: "image/webp" });
/** A 16 by 8 px WebP, which the server refuses as not square. */
const wide = new Blob([readFileSync(new URL("../testdata/wide.webp", import.meta.url))], { type: "image/webp" });

/** Returns the error a promise rejects with, or fails when it resolves. */
async function rejection(promise: Promise<unknown>): Promise<unknown> {
  return await promise.then(
    () => expect.fail("the request succeeded"),
    (error: unknown) => error
  );
}

describe("profileGetSelf", () => {
  it("returns 404 before the first write", async () => {
    await expectStatus(profileGetSelf(api, await userAccessToken()), 404);
  });

  it("returns the caller's profile", async () => {
    const accessToken = await userAccessToken();
    const profile = await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });

    await expect(profileGetSelf(api, accessToken)).resolves.toEqual(profile);
  });

  it("returns 403 for an anonymous session", async () => {
    await expectStatus(profileGetSelf(api, await anonymousAccessToken()), 403);
  });
});

describe("profileUsernameUpdate", () => {
  it("creates the profile on the first write, with the username trimmed", async () => {
    const profile = await profileUsernameUpdate(api, await userAccessToken(), { username: "  Jean Paul  " });

    expect(profile).toEqual({ id: expect.any(String), username: "Jean Paul" });
  });

  it("replaces the username of the profile", async () => {
    const accessToken = await userAccessToken();
    const created = await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });

    await expect(profileUsernameUpdate(api, accessToken, { username: "Marie" })).resolves.toEqual({
      id: created.id,
      username: "Marie",
    });
  });

  it("clears the username with a name of spaces", async () => {
    const accessToken = await userAccessToken();
    const created = await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });

    await expect(profileUsernameUpdate(api, accessToken, { username: "  " })).resolves.toEqual({ id: created.id });
  });

  it("returns 400 naming the username for a refused one", async () => {
    const error = await rejection(profileUsernameUpdate(api, await userAccessToken(), { username: "Jean\tPaul" }));

    expect(isHttpError(error) && error.status).toBe(400);
    expect(isHttpError(error) && error.tags.invalidFields).toEqual({ username: "username" });
  });

  it("returns 403 for an anonymous session", async () => {
    await expectStatus(profileUsernameUpdate(api, await anonymousAccessToken(), { username: "Jean Paul" }), 403);
  });
});

describe("profilePictureUpdate", () => {
  it("creates the profile on the first write, with a picture version", async () => {
    const profile = await profilePictureUpdate(api, await userAccessToken(), { picture: square });

    expect(profile).toEqual({ id: expect.any(String), pictureVersion: expect.any(String) });
  });

  it("replaces the picture with a new version", async () => {
    const accessToken = await userAccessToken();
    const first = await profilePictureUpdate(api, accessToken, { picture: square });
    const second = await profilePictureUpdate(api, accessToken, { picture: square });

    expect(second.id).toBe(first.id);
    expect(second.pictureVersion).not.toBe(first.pictureVersion);
  });

  it("returns 422 with the reason for a refused picture", async () => {
    const accessToken = await userAccessToken();
    const refusals = [
      { picture: wide, reason: PictureRefusal.Ratio },
      { picture: new Blob(["not an image"], { type: "image/webp" }), reason: PictureRefusal.Format },
    ];

    for (const { picture, reason } of refusals) {
      const error = await rejection(profilePictureUpdate(api, accessToken, { picture }));

      expect(isHttpError(error) && error.status).toBe(422);
      expect(isHttpError(error) && error.tags.reason).toBe(reason);
    }
  });

  it("returns 413 above 256 KiB", async () => {
    const picture = new Blob([new Uint8Array(256 * 1024 + 1)], { type: "image/webp" });

    await expectStatus(profilePictureUpdate(api, await userAccessToken(), { picture }), 413);
  });

  it("returns 403 for an anonymous session", async () => {
    await expectStatus(profilePictureUpdate(api, await anonymousAccessToken(), { picture: square }), 403);
  });
});

describe("profilePictureDelete", () => {
  it("removes the picture", async () => {
    const accessToken = await userAccessToken();
    const created = await profilePictureUpdate(api, accessToken, { picture: square });

    await expect(profilePictureDelete(api, accessToken)).resolves.toEqual({ id: created.id });
    await expectStatus(profilePictureGet(api, accessToken, { id: created.id }), 404);
  });

  it("returns 404 before the first write", async () => {
    await expectStatus(profilePictureDelete(api, await userAccessToken()), 404);
  });

  it("returns 403 for an anonymous session", async () => {
    await expectStatus(profilePictureDelete(api, await anonymousAccessToken()), 403);
  });
});

describe("profileGet", () => {
  it("reads any profile with an anonymous token", async () => {
    const accessToken = await userAccessToken();
    await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });
    const profile = await profilePictureUpdate(api, accessToken, { picture: square });

    await expect(profileGet(api, await anonymousAccessToken(), { id: profile.id })).resolves.toEqual(profile);
  });

  it("never names the owner", async () => {
    const accessToken = await userAccessToken();
    await profileUsernameUpdate(api, accessToken, { username: "Jean Paul" });
    const { id } = await profilePictureUpdate(api, accessToken, { picture: square });

    // The client's schema drops unknown fields, so the raw body proves the response holds no other.
    const response = await fetch(`${process.env.REST_URL}/v1/profile?id=${id}`, {
      headers: { Authorization: `Bearer ${await anonymousAccessToken()}` },
    });

    expect(Object.keys(await response.json()).sort()).toEqual(["id", "pictureVersion", "username"]);
  });

  it("returns 404 for an unknown profile", async () => {
    await expectStatus(profileGet(api, await anonymousAccessToken(), { id: crypto.randomUUID() }), 404);
  });

  it("returns 400 for a malformed ID", async () => {
    await expectStatus(profileGet(api, await anonymousAccessToken(), { id: "not-a-uuid" }), 400);
  });
});

describe("profilePictureGet", () => {
  it("reads any picture with an anonymous token, as it was sent", async () => {
    const profile = await profilePictureUpdate(api, await userAccessToken(), { picture: square });
    const picture = await profilePictureGet(api, await anonymousAccessToken(), { id: profile.id });

    expect(picture.type).toBe("image/webp");
    expect(new Uint8Array(await picture.arrayBuffer())).toEqual(new Uint8Array(await square.arrayBuffer()));
  });

  it("returns 404 for a profile without a picture", async () => {
    const profile = await profileUsernameUpdate(api, await userAccessToken(), { username: "Jean Paul" });

    await expectStatus(profilePictureGet(api, await anonymousAccessToken(), { id: profile.id }), 404);
  });

  it("returns 400 for a malformed ID", async () => {
    await expectStatus(profilePictureGet(api, await anonymousAccessToken(), { id: "not-a-uuid" }), 400);
  });
});
