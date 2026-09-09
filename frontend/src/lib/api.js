export async function api(path, options = {}) {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, {
    credentials: "include",
    ...options,
    headers,
  });
  if (response.status === 204) return {};
  let body;
  try {
    body = await response.json();
  } catch (error) {
    if (error.name === "AbortError") throw error;
    const failure = new Error(response.ok ? "服务器返回了无效数据，请刷新重试" : "请求失败，请稍后重试");
    failure.status = response.status;
    throw failure;
  }
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    const error = new Error("服务器返回了无效数据，请刷新重试");
    error.status = response.status;
    throw error;
  }
  if (!response.ok) { const error = new Error(body.error || "请求失败，请稍后重试"); error.status = response.status; throw error; }
  return body;
}
