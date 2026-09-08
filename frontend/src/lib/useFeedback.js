import { useRef, useState } from "react";

export function useFeedback() {
  const [notice, setNotice] = useState(null);
  const [busy, setBusy] = useState(false);
  const pending = useRef(false);
  function fail(error) { setNotice({ type: "error", text: error.message }); }
  async function run(action) {
    if (pending.current) return;
    pending.current = true;
    setBusy(true);
    setNotice(null);
    try {
      await action();
      setNotice({ type: "success", text: "操作成功" });
    } catch (error) {
      fail(error);
    } finally {
      pending.current = false;
      setBusy(false);
    }
  }
  return { notice, busy, run, fail };
}
