export function venueDate(date = new Date()) {
  return new Date(date.getTime() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10);
}

export function venueTime(value) {
  return new Date(new Date(value).getTime() + 8 * 60 * 60 * 1000).toISOString().slice(11, 16);
}

export function slotPeriod(slot) {
  const hour = Number(venueTime(slot.startsAt).slice(0, 2));
  return hour < 12 ? "morning" : hour < 18 ? "afternoon" : "evening";
}

export function groupBookings(items) {
  return {
    upcoming: items.filter((item) => item.status === "confirmed").sort((a, b) => new Date(a.startsAt) - new Date(b.startsAt)),
    history: items.filter((item) => item.status !== "confirmed").sort((a, b) => new Date(b.startsAt) - new Date(a.startsAt)),
  };
}
