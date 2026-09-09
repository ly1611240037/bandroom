import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { CalendarDays, Clock3, DoorOpen, Package, Search, ShieldCheck } from "lucide-react";
import { api } from "../lib/api";
import { filterRooms } from "../lib/rooms.js";
import { equipmentQuantity } from "../lib/equipment.js";
import RoomScene from "../components/RoomScene";
import { Card, Empty, Loading, Notice, PageHeader, SectionTitle, Status } from "../components/ui";

export default function Home() {
  const [data, setData] = useState(null);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  const [people, setPeople] = useState("");
  const [keyword, setKeyword] = useState("");
  const navigate = useNavigate();
  useEffect(() => {
    const controller = new AbortController();
    setError("");
    Promise.all(["/api/rooms", "/api/equipment", "/api/venue"].map((path) => api(path, { signal: controller.signal })))
      .then(([r, e, v]) => { if (!controller.signal.aborted) setData({ rooms: r.rooms || [], equipment: e.equipment || [], venue: Object.fromEntries((v.content || []).map((x) => [x.key, x.value])) }); })
      .catch((e) => { if (!controller.signal.aborted) setError(e.message); });
    return () => controller.abort();
  }, [revision]);
  if (error) return <><Notice type="error">{error}</Notice><button className="button ghost" onClick={() => setRevision((value) => value + 1)}>重新加载场馆</button></>;
  if (!data) return <Loading />;
  const rooms = filterRooms(data.rooms, people, keyword);
  function clear() { setPeople(""); setKeyword(""); }
  return <>
    <PageHeader eyebrow="CUSTOMER SPACE" title="欢迎来到 BandRoom" description="先了解场馆，再选择适合乐队的排练时段。" action={<button className="button primary" onClick={() => navigate("/customer/bookings")}><CalendarDays size={17} />立即预约</button>} />
    <div className="hero-card"><div><span className="eyebrow">{data.venue.name || "BANDROOM MUSIC SPACE"}</span><h2>{data.venue.description || "为每一次排练，准备好合适的空间。"}</h2><p>{data.venue.announcement || "支持多房间预约、会员卡和额外设备借用。"}</p></div><div className="hero-note"><Clock3 size={28} /><span>营业说明</span><b>{data.venue.openingRules || "08:00 — 23:00"}</b></div></div>
    <Card className="room-finder">
      <SectionTitle icon={Search} title="找合适的排练房" />
      <div className="room-finder-fields">
        <label className="field"><span>排练人数</span><input type="number" min="1" step="1" placeholder="不限人数" value={people} onChange={(e) => setPeople(e.target.value)} /></label>
        <label className="field"><span>房间或固定设备</span><input type="search" placeholder="例如：架子鼓、钢琴、A 房" value={keyword} onChange={(e) => setKeyword(e.target.value)} /></label>
        <button className="button ghost" disabled={!people && !keyword} onClick={clear}>清除筛选</button>
      </div>
      <p className="form-hint" role="status">找到 {rooms.length} / {data.rooms.length} 间房 · 按房间容量及可用固定设备匹配，空闲时段请进入预约查看。</p>
    </Card>
    {rooms.length ? <div className="room-scenes">{rooms.map((room) => <RoomScene compact key={room.id} room={room} accent={["#ffb466", "#87baff", "#87dbc0"][data.rooms.indexOf(room) % 3]} onBook={(id) => navigate(`/customer/bookings?roomId=${id}`)} />)}</div> : <Card><Empty icon={DoorOpen} title={data.rooms.length ? "没有符合条件的房间" : "暂无房间资料"} description={data.rooms.length ? "试试减少人数，或更换设备关键词。" : "门店发布房间后会显示在这里。"} /></Card>}
    <div className="info-grid venue-extras"><Card><SectionTitle icon={Package} title={`公共借用设备 · ${data.equipment.length} 类`} /><p className="form-hint">固定设备已包含在房间内，以下设备需预约时另选；数量为总库存。</p><div className="public-equipment-grid">{data.equipment.map((item) => <div className="public-equipment-card" key={item.id}><div><b>{item.name}</b><Status>{item.status}</Status></div><p>{item.description || "设备说明待补充"}</p><strong>{equipmentQuantity(item)}</strong></div>)}</div>{!data.equipment.length && <Empty title="暂无公共设备" />}</Card><Card><SectionTitle icon={ShieldCheck} title="会员说明" /><p className="card-copy">{data.venue.membershipInstructions || "月卡或次数卡由老板线下开通，预约时直接使用有效权益。"}</p></Card></div>
  </>;
}
