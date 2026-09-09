import { DoorOpen, Drum, Guitar, Piano, Speaker, Music2, UsersRound } from "lucide-react";
import { equipmentQuantity } from "../lib/equipment.js";
import RoomLayout from "./RoomLayout";
import { Status } from "./ui";

function EquipmentIcon({ name }) {
  const Icon = name.includes("鼓") ? Drum : name.includes("钢琴") ? Piano : name.includes("吉他") || name.includes("贝斯") ? Guitar : name.includes("扩声") ? Speaker : Music2;
  return <Icon size={20} />;
}

export default function RoomScene({ room, onBook, compact = false }) {
  const equipment = room.fixedEquipment || [];
  const details = <>
    {compact && <p className="card-copy">{room.description || "房间介绍待补充。"}</p>}
    <RoomLayout room={room} />
    <p className="form-hint">设备分区示意，非实景或精确比例。固定设备随房使用；公共设备需在预约时另选。</p>
    <ul className="room-equipment">{equipment.map((item, index) => <li key={item.id}><span><b>{index + 1}. {item.name} · {equipmentQuantity(item)}</b><small>{item.description}</small></span><Status>{item.status}</Status></li>)}</ul>
  </>;
  return <article className={`room-scene ${compact ? "room-preview" : ""}`}>
    {compact && <div className="room-preview-banner" aria-hidden="true"><DoorOpen size={38} /><span>REHEARSAL ROOM</span></div>}
    <div className="room-scene-heading"><div><h3>{room.name}</h3><span><UsersRound size={14} /> {room.capacity ? `${room.capacity} 人以内` : "容量请咨询门店"}</span></div><Status>{room.status}</Status></div>
    <p className="card-copy room-intro">{(compact ? room.description?.split("。")[0] : room.description) || "房间介绍待补充。"}</p>
    {compact ? <><div className="room-preview-equipment"><span className="muted">随房设备 · {equipment.length} 类</span><div className="room-equipment-tags">{equipment.map((item) => <span key={item.id} className={item.status !== "available" ? "unavailable" : ""}><EquipmentIcon name={item.name} />{item.name} · {equipmentQuantity(item)}{item.status !== "available" && "（暂不可用）"}</span>)}</div>{!equipment.length && <p className="muted">设备资料待补充</p>}</div><details className="room-preview-details"><summary>查看布局与设备详情</summary>{details}</details></> : details}
    {onBook && <div className="room-preview-action"><button type="button" className="button primary full" disabled={room.status !== "available"} onClick={() => onBook(room.id)}>{room.status === "available" ? "预约这间房" : "暂不可预约"}</button></div>}
  </article>;
}
