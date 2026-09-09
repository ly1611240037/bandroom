export async function api(path, options = {}) {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, {
    credentials: "include",
    ...options,
    headers,
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) { const error = new Error(body.error || "请求失败，请稍后重试"); error.status = response.status; throw error; }
  return body;
}
