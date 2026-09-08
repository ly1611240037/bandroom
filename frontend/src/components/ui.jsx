import { AlertCircle, CheckCircle2, ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";

export function Notice({ children, type = "success" }) {
  return <div className={`notice ${type}`}><span>{type === "error" ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}</span>{children}</div>;
}
export function PageHeader({ eyebrow, title, description, action }) {
  return <header className="page-header"><div><span className="eyebrow">{eyebrow}</span><h1>{title}</h1>{description && <p>{description}</p>}</div>{action}</header>;
}
export function Card({ children, className = "" }) { return <section className={`surface-card ${className}`}>{children}</section>; }
export function Status({ children, tone = "green" }) { return <span className={`status ${tone}`}>{children}</span>; }
export function Empty({ icon: Icon = AlertCircle, title = "暂无内容", description = "这里还没有记录。" }) { return <div className="empty-state"><Icon size={30} /><b>{title}</b><span>{description}</span></div>; }
export function SectionTitle({ icon: Icon, title, link }) { return <div className="section-title"><span><Icon size={18} />{title}</span>{link && <Link to={link}>查看全部 <ChevronRight size={15} /></Link>}</div>; }
