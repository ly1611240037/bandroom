import { useCallback, useEffect, useRef, useState } from "react";
import { Bell, CheckCircle2, RefreshCw } from "lucide-react";
import { api } from "../lib/api";
import { venueDate, venueTime } from "../lib/date.js";
import { Card, Empty, Loading, Notice, PageHeader } from "../components/ui";

function timestamp(value) {
  // SQLite CURRENT_TIMESTAMP is UTC without a timezone suffix.
  const date = new Date(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(value) ? value.replace(" ", "T") + "Z" : value);
  return Number.isNaN(date.getTime()) ? "时间未提供" : `${venueDate(date)} ${venueTime(date)}`;
}

export default function Notifications() {
  const [items, setItems] = useState(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState(null);
  const request = useRef(null);
  const pending = useRef(false);
  const refresh = useCallback(async () => {
    if (pending.current) return;
    request.current?.abort();
    const controller = new AbortController(); request.current = controller;
    setLoading(true); setNotice(null);
    try {
      const data = await api("/api/notifications", { signal: controller.signal });
      if (!controller.signal.aborted) setItems(data.notifications || []);
    } catch (error) {
      if (!controller.signal.aborted) setNotice({ type: "error", text: error.message });
    } finally {
      if (!controller.signal.aborted) setLoading(false);
    }
  }, []);
  useEffect(() => { refresh(); return () => request.current?.abort(); }, [refresh]);
  async function read(id) {
    if (pending.current || loading) return;
    pending.current = true; setBusy(true); setNotice(null);
    const controller = new AbortController(); request.current = controller;
    const throughId = (items || []).reduce((max, item) => Math.max(max, item.id), 0);
    try {
      await api(id ? `/api/notifications/${id}/read` : "/api/notifications/read-all", {
        method: "PATCH", signal: controller.signal, ...(id ? {} : { body: JSON.stringify({ throughId }) }),
      });
      if (!controller.signal.aborted) {
        setItems((old) => old.map((item) => !item.readAt && (id ? item.id === id : item.id <= throughId) ? { ...item, readAt: new Date().toISOString() } : item));
        setNotice({ type: "success", text: id ? "通知已标记为已读" : "已加载的通知已全部标记为已读" });
      }
    } catch (error) {
      if (!controller.signal.aborted) setNotice({ type: "error", text: error.message });
    } finally {
      pending.current = false;
      if (!controller.signal.aborted) setBusy(false);
    }
  }
  const unread = (items || []).filter((item) => !item.readAt);
  return <>
    <PageHeader eyebrow="NOTIFICATIONS" title="通知" description="查看预约与会员消息，时间均为北京时间。" action={<div className="notification-actions"><button className="button ghost" disabled={busy || loading} onClick={refresh}><RefreshCw size={16} />{notice?.type === "error" ? "重新加载" : "刷新通知"}</button><button className="button primary" disabled={busy || loading || !unread.length} onClick={() => read()}><CheckCircle2 size={16} />{busy ? "处理中…" : "全部标为已读"}</button></div>} />
    {notice && <Notice type={notice.type}>{notice.text}</Notice>}
    {loading && <Loading />}
    {items !== null && [false, true].map((isRead) => {
      const group = items.filter((item) => !!item.readAt === isRead);
      return <section className="page-section" key={String(isRead)}><div className="section-heading"><h2>{isRead ? "已读通知" : "未读通知"}</h2><span>{group.length} 条</span></div><Card className="notification-list">
        {!group.length && <Empty icon={isRead ? Bell : CheckCircle2} title={isRead ? "暂无已读通知" : "没有未读通知"} description={isRead ? "读过的通知会保留在这里。" : "新消息会在这里显示，可点击刷新查看。"} />}
        {group.map((item) => {
          const content = <><Bell size={17} /><span><b>{item.title}</b><small>{item.body}</small><small className="notification-time">{timestamp(item.createdAt)} · {isRead ? "已读" : "点击标为已读"}</small></span></>;
          return isRead ? <div className="notification-row read" key={item.id}>{content}</div> : <button key={item.id} disabled={busy || loading} className="notification-row" onClick={() => read(item.id)}>{content}</button>;
        })}
      </Card></section>;
    })}
  </>;
}
