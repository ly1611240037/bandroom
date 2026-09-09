import { useId } from "react";

// Isometric projection: floor coordinates stay separate from equipment height.
function project(x, y, z = 0) { return [240 + (x - y) * .88, 145 + (x + y) * .43 - z]; }
function points(vertices) { return vertices.map(([x, y, z]) => project(x, y, z).join(",")).join(" "); }

function EquipmentModel({ item, index, x, y, size }) {
  const drum = item.name.includes("鼓");
  const piano = item.name.includes("钢琴");
  const height = drum ? 23 : piano ? 20 : 36;
  const width = piano ? size * 1.25 : size;
  const depth = piano ? size * .55 : size;
  const [cx, cy] = project(x + width / 2, y + depth / 2, height);
  return <g className={`iso-device ${item.status !== "available" ? "unavailable" : ""}`}>
    <title>{index + 1}. {item.name} × {item.quantity}，{item.status === "available" ? "随房可用" : "暂不可用"}</title>
    <polygon points={points([[x + 6, y + 6, 0], [x + width + 9, y + 6, 0], [x + width + 9, y + depth + 9, 0], [x + 6, y + depth + 9, 0]])} fill="#0005" />
    {drum ? <>
      <path d={`M ${cx - 18} ${cy} v 22 a 18 9 0 0 0 36 0 v -22`} fill="#95533d" stroke="#e3a87e" />
      <ellipse cx={cx} cy={cy} rx="18" ry="9" fill="#e9d8bb" stroke="#ac815d" strokeWidth="3" />
      <path d={`M ${cx + 25} ${cy - 10} v 34 m -8 4 l 8 -4 l 8 4`} stroke="#94a3b8" fill="none" strokeWidth="2" />
      <ellipse cx={cx + 25} cy={cy - 10} rx="14" ry="5" fill="#e3b66c" />
    </> : <>
      <polygon points={points([[x, y + depth, 0], [x + width, y + depth, 0], [x + width, y + depth, height], [x, y + depth, height]])} fill="#293953" stroke="#7390ab" />
      <polygon points={points([[x + width, y, 0], [x + width, y + depth, 0], [x + width, y + depth, height], [x + width, y, height]])} fill="#142033" stroke="#7390ab" />
      <polygon points={points([[x, y, height], [x + width, y, height], [x + width, y + depth, height], [x, y + depth, height]])} fill={piano ? "#e4e8eb" : "#526980"} stroke="#9fb3c4" />
      {piano ? Array.from({ length: 8 }, (_, key) => <polyline key={key} points={points([[x + key * width / 8, y, height + 1], [x + key * width / 8, y + depth, height + 1]])} stroke="#293953" />) : <ellipse cx={project(x + width / 2, y + depth, height / 2)[0]} cy={project(x + width / 2, y + depth, height / 2)[1]} rx="8" ry="11" fill="#0a1220" stroke="#697b91" />}
    </>}
    <circle cx={cx} cy={cy - 23} r="11" fill={item.status === "available" ? "#ffb466" : "#64748b"} stroke="#101b2e" strokeWidth="2" />
    <text x={cx} y={cy - 19} textAnchor="middle" fontSize="11" fontWeight="700" fill="#111827">{index + 1}</text>
  </g>;
}

export default function RoomLayout({ room }) {
  const id = useId();
  const equipment = room.fixedEquipment || [];
  const columns = Math.max(2, Math.ceil(Math.sqrt(equipment.length)));
  const step = 180 / columns;
  return <figure className="room-layout-3d">
    <svg viewBox="0 0 480 385" role="img" aria-labelledby={`${id}-title ${id}-description`}>
      <title id={`${id}-title`}>{room.name}立体布局示意</title>
      <desc id={`${id}-description`}>等距视角展示房间地板、两面墙和设备。数字对应下方设备清单，非实景与精确比例。</desc>
      <defs><linearGradient id={`${id}-floor`} x2="0" y2="1"><stop stopColor="#b18159" /><stop offset="1" stopColor="#624833" /></linearGradient></defs>
      <ellipse cx="240" cy="331" rx="203" ry="31" fill="#0004" />
      <polygon points={points([[0, 0, 0], [220, 0, 0], [220, 220, 0], [0, 220, 0]])} fill={`url(#${id}-floor)`} stroke="#d1a678" strokeWidth="2" />
      {Array.from({ length: 10 }, (_, i) => <polyline key={i} points={points([[i * 22, 0, 1], [i * 22, 220, 1]])} stroke="#dec097" opacity=".2" />)}
      <polygon points={points([[0, 0, 0], [0, 220, 0], [0, 220, 100], [0, 0, 100]])} fill="#243852" stroke="#58728b" strokeWidth="2" />
      <polygon points={points([[0, 0, 0], [220, 0, 0], [220, 0, 100], [0, 0, 100]])} fill="#344b67" stroke="#7390aa" strokeWidth="2" />
      {[35, 85, 135, 185].map((x) => <g key={x}><polygon points={points([[x, 0, 28], [x + 25, 0, 28], [x + 25, 0, 80], [x, 0, 80]])} fill="#18283e" stroke="#49617c" /><polygon points={points([[0, x, 28], [0, x + 25, 28], [0, x + 25, 80], [0, x, 80]])} fill="#132235" stroke="#3a536e" /></g>)}
      <polygon points={points([[45, 85, 1], [165, 85, 1], [165, 180, 1], [45, 180, 1]])} fill="#263b4d" opacity=".65" stroke="#b6aa91" strokeDasharray="4 3" />
      {equipment.map((item, index) => <EquipmentModel key={item.id} item={item} index={index} x={20 + index % columns * step} y={20 + Math.floor(index / columns) * step} size={Math.min(35, step * .55)} />)}
      <text x="240" y="365" textAnchor="middle" fill="#aebed0" fontSize="12">开放视角 · 前侧入口</text>
    </svg>
    <figcaption>立体布局示意 · 编号对应设备清单</figcaption>
  </figure>;
}
