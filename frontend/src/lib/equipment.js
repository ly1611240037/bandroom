export function isDrumKit(name) { return name.includes("架子鼓"); }
export function equipmentQuantity(item) {
  return `${item.quantity} ${isDrumKit(item.name) || item.name.includes("系统") ? "套" : item.name.includes("镲") ? "片" : "件"}`;
}
