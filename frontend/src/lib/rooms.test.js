import test from "node:test";
import assert from "node:assert/strict";
import { filterRooms } from "./rooms.js";

test("room finder combines capacity and available equipment without matching maintenance inventory", () => {
  const rooms = [
    { name: "A 房", capacity: 6, fixedEquipment: [{ name: "架子鼓", quantity: 1, status: "available" }] },
    { name: "B 房", capacity: 8, fixedEquipment: [{ name: "架子鼓", quantity: 1, status: "maintenance" }] },
    { name: "C 房", capacity: 4 },
  ];
  assert.deepEqual(filterRooms(rooms, "", ""), rooms);
  assert.deepEqual(filterRooms(rooms, 5, " 鼓 "), [rooms[0]]);
  assert.deepEqual(filterRooms(rooms, 7, "鼓"), []);
  assert.deepEqual(filterRooms(rooms, "", "a 房"), [rooms[0]]);
});
