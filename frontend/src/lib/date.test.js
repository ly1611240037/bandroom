import test from "node:test";
import assert from "node:assert/strict";
import { venueDate, venueTime, slotPeriod, groupBookings } from "./date.js";

test("booking times and date use venue time across UTC midnight", () => {
  assert.equal(venueDate(new Date("2026-09-08T18:30:00Z")), "2026-09-09");
  assert.equal(venueTime("2026-09-08T18:30:00Z"), "02:30");
  assert.equal(venueTime("2026-09-09T02:30:00+08:00"), "02:30");
});

test("time ranges include every slot exactly once at noon and evening boundaries", () => {
  const slots = ["11:30", "12:00", "17:30", "18:00"].map((time) => ({ startsAt: `2026-09-09T${time}:00+08:00` }));
  assert.deepEqual(slots.map(slotPeriod), ["morning", "afternoon", "afternoon", "evening"]);
});

test("history uses server status and sorts instants without mutating input", () => {
  const items = [
    { id: 1, status: "completed", startsAt: "2026-09-09T09:00:00+08:00" },
    { id: 2, status: "confirmed", startsAt: "2026-09-10T09:00:00+08:00" },
    { id: 3, status: "cancelled", startsAt: "2026-09-09T02:00:00Z" },
    { id: 4, status: "confirmed", startsAt: "2026-09-09T08:00:00+08:00" },
  ];
  const { upcoming, history } = groupBookings(items);
  assert.deepEqual(upcoming.map((item) => item.id), [4, 2]);
  assert.deepEqual(history.map((item) => item.id), [3, 1]);
  assert.deepEqual(items.map((item) => item.id), [1, 2, 3, 4]);
});
