import type { ProfileApi } from "./api";

import { HTTP_HEADERS } from "@a-novel-kit/nodelib-browser/http";

import { z } from "zod";

/** The rules a picture breaks when the server refuses it, as a 422 error's `tags.reason` names them. */
export enum PictureRefusal {
  /** The bytes are neither a still WebP nor a JPEG, or do not decode whole. */
  Format = "format",
  /** The picture is not square. */
  Ratio = "ratio",
  /** The picture's side is under 1 or over 512 pixels. */
  Dimensions = "dimensions",
}

/** A user's public profile. It never names the user who owns it. */
export const ProfileSchema = z.object({
  id: z.uuid(),
  /** The name the profile shows, absent until its owner sets one. */
  username: z.string().optional(),
  /**
   * When the picture last changed, absent while the profile has no picture. A picture URL that carries it
   * names one version, which a cache keeps for good.
   */
  pictureVersion: z.iso.datetime({ offset: true }).optional(),
});

/** The value {@link ProfileSchema} validates. */
export type Profile = z.infer<typeof ProfileSchema>;

/** The profile to read. */
export const ProfileGetRequestSchema = z.object({
  id: z.uuid(),
});

/** The value {@link ProfileGetRequestSchema} validates. */
export type ProfileGetRequest = z.infer<typeof ProfileGetRequestSchema>;

/** The caller's new username. The server trims the spaces at both ends; a name left empty clears it. */
export const ProfileUsernameUpdateRequestSchema = z.object({
  username: z.string(),
});

/** The value {@link ProfileUsernameUpdateRequestSchema} validates. */
export type ProfileUsernameUpdateRequest = z.infer<typeof ProfileUsernameUpdateRequestSchema>;

/** The caller's new picture: an encoded still WebP or JPEG, square, of 1 to 512 pixels and 256 KiB at most. */
export const ProfilePictureUpdateRequestSchema = z.object({
  picture: z.instanceof(Blob),
});

/** The value {@link ProfilePictureUpdateRequestSchema} validates. */
export type ProfilePictureUpdateRequest = z.infer<typeof ProfilePictureUpdateRequestSchema>;

/** The profile whose picture to read. */
export const ProfilePictureGetRequestSchema = z.object({
  id: z.uuid(),
});

/** The value {@link ProfilePictureGetRequestSchema} validates. */
export type ProfilePictureGetRequest = z.infer<typeof ProfilePictureGetRequestSchema>;

/**
 * Returns the caller's own profile. Throws on HTTP 404 before the caller's first write, and on 403 for an
 * anonymous session.
 */
export async function profileGetSelf(api: ProfileApi, accessToken: string): Promise<Profile> {
  return await api.fetch("/v1/profile/self", ProfileSchema, {
    method: "GET",
    headers: { ...HTTP_HEADERS.JSON, Authorization: `Bearer ${accessToken}` },
  });
}

/** Returns any profile. Any access token reads it, an anonymous one included. Throws on HTTP 404 for an unknown ID. */
export async function profileGet(api: ProfileApi, accessToken: string, request: ProfileGetRequest): Promise<Profile> {
  const params = new URLSearchParams({ id: request.id });
  return await api.fetch(`/v1/profile?${params.toString()}`, ProfileSchema, {
    method: "GET",
    headers: { ...HTTP_HEADERS.JSON, Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Sets or clears the caller's username, and returns the caller's profile, which the first write creates. Throws on HTTP
 * 400, with `username` in the error's `tags.invalidFields`, for a name the server refuses, and on 403 for an
 * anonymous session.
 */
export async function profileUsernameUpdate(
  api: ProfileApi,
  accessToken: string,
  request: ProfileUsernameUpdateRequest
): Promise<Profile> {
  return await api.fetch("/v1/profile", ProfileSchema, {
    method: "PUT",
    headers: { ...HTTP_HEADERS.JSON, Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(request),
  });
}

/**
 * Replaces the caller's picture, and returns the caller's profile, which the first write creates. The server
 * reads the format from the bytes. Throws on HTTP 422, with a {@link PictureRefusal} in the error's
 * `tags.reason`, for a picture it refuses, on 413 above 256 KiB, and on 403 for an anonymous session.
 */
export async function profilePictureUpdate(
  api: ProfileApi,
  accessToken: string,
  request: ProfilePictureUpdateRequest
): Promise<Profile> {
  return await api.fetch("/v1/profile/picture", ProfileSchema, {
    method: "PUT",
    headers: { "Content-Type": request.picture.type, Authorization: `Bearer ${accessToken}` },
    body: request.picture,
  });
}

/**
 * Removes the caller's picture, and returns the caller's profile. Throws on HTTP 404 before the caller's first
 * write, and on 403 for an anonymous session.
 */
export async function profilePictureDelete(api: ProfileApi, accessToken: string): Promise<Profile> {
  return await api.fetch("/v1/profile/picture", ProfileSchema, {
    method: "DELETE",
    headers: { ...HTTP_HEADERS.JSON, Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Returns the picture of any profile, typed with the media type it was stored with. Any access token reads it,
 * an anonymous one included. Throws on HTTP 404 when the profile does not exist or holds no picture.
 */
export async function profilePictureGet(
  api: ProfileApi,
  accessToken: string,
  request: ProfilePictureGetRequest
): Promise<Blob> {
  const params = new URLSearchParams({ id: request.id });
  return await api.fetchBlob(`/v1/profile/picture?${params.toString()}`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}
