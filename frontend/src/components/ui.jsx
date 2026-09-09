import { AlertCircle, CheckCircle2, ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";

export function Notice({ children, type = "success" }) {
  return <div role={type === "error" ? "alert" : "status"} className={`notice ${type}`}><span>{type === "error" ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}</span>{children}</div>;
}
export function PageHeader({ eyebrow, title, description, action }) {
  return <header className="page-header"><div><span className="eyebrow">{eyebrow}</span><h1>{title}</h1>{description && <p>{description}</p>}</div>{action}</header>;
}
export function Card({ children, className = "" }) { return <section className={`surface-card ${className}`}>{children}</section>; }
const statuses = { confirmed: ["已确认", "green"], completed: ["已完成", "gray"], cancelled: ["已取消", "gray"], no_show: ["未到场", "gray"], available: ["可用", "green"], maintenance: ["维护中", "gray"], disabled: ["已停用", "gray"], suspended: ["暂停预约", "gray"], active: ["当前可用", "green"], scheduled: ["待生效", "gray"], expired: ["已过期", "gray"], depleted: ["次数已用完", "gray"] };
export function Status({ children, tone }) { const status = statuses[children]; return <span className={`status ${tone || status?.[1] || "gray"}`}>{status?.[0] || children}</span>; }
export function Loading() { return <p className="loading-state" role="status">正在加载…</p>; }
export function Empty({ icon: Icon = AlertCircle, title = "暂无内容", description = "这里还没有记录。" }) { return <div className="empty-state"><Icon size={30} /><b>{title}</b><span>{description}</span></div>; }
export function SectionTitle({ icon: Icon, title, link }) { return <div className="section-title"><span><Icon size={18} />{title}</span>{link && <Link to={link}>查看全部 <ChevronRight size={15} /></Link>}</div>; }
