import { test } from "node:test";
import assert from "node:assert/strict";
import { api } from "./api.js";
import { venueDate } from "./date.js";

test("merges custom Headers without losing JSON content type or credentials", async (t) => {
  t.mock.method(globalThis, "fetch", async (_path, options) => {
    assert.equal(options.credentials, "include");
    assert.equal(options.headers.get("Content-Type"), "application/json");
    assert.equal(options.headers.get("X-Request-ID"), "test");
    return Response.json({ ok: true });
  });
  assert.deepEqual(await api("/api/test", { headers: new Headers({ "X-Request-ID": "test" }) }), { ok: true });
});

test("preserves explicit content type and abort signal", async (t) => {
  const controller = new AbortController();
  t.mock.method(globalThis, "fetch", async (_path, options) => {
    assert.equal(options.headers.get("Content-Type"), "text/plain");
    assert.equal(options.signal, controller.signal);
    return new Response(null, { status: 204 });
  });
  assert.deepEqual(await api("/api/test", { headers: { "content-type": "text/plain" }, signal: controller.signal }), {});
});

test("exposes business errors and tolerates non-JSON error responses", async (t) => {
  t.mock.method(globalThis, "fetch", async () => Response.json({ error: "时段已占用" }, { status: 409 }));
  await assert.rejects(api("/api/test"), /时段已占用/);
  t.mock.method(globalThis, "fetch", async () => new Response("Bad Gateway", { status: 502 }));
  await assert.rejects(api("/api/test"), /请求失败/);
});

test("venue date follows UTC+8 at midnight regardless of browser timezone", () => {
  assert.equal(venueDate(new Date("2026-09-08T15:59:59Z")), "2026-09-08");
  assert.equal(venueDate(new Date("2026-09-08T16:00:00Z")), "2026-09-09");
});

test("JSON API distinguishes valid data, empty responses and malformed success", async (t) => {
  const fetch = t.mock.method(globalThis, "fetch");
  for (const value of ["<html>fallback</html>", "null", "[]", "\"ok\"", ""]) {
    fetch.mock.mockImplementation(async () => new Response(value));
    await assert.rejects(api("/api/test"), { message: "服务器返回了无效数据，请刷新重试", status: 200 });
  }
  fetch.mock.mockImplementation(async () => new Response('{"rooms":[]}'));
  assert.deepEqual(await api("/api/test"), { rooms: [] });
  fetch.mock.mockImplementation(async () => new Response(null, { status: 204 }));
  assert.deepEqual(await api("/api/test"), {});
});

test("JSON API preserves HTTP status, server errors and body cancellation", async (t) => {
  const fetch = t.mock.method(globalThis, "fetch", async () => new Response('{"error":"请先登录"}', { status: 401 }));
  await assert.rejects(api("/api/test"), { message: "请先登录", status: 401 });
  fetch.mock.mockImplementation(async () => new Response("Bad Gateway", { status: 502 }));
  await assert.rejects(api("/api/test"), { status: 502 });
  const cancelled = new DOMException("cancelled", "AbortError");
  fetch.mock.mockImplementation(async () => ({ status: 200, ok: true, json: async () => { throw cancelled; } }));
  await assert.rejects(api("/api/test"), (error) => error === cancelled);
});
