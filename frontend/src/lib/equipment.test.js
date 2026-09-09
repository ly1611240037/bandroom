import test from "node:test";
import assert from "node:assert/strict";
import { isDrumKit, equipmentQuantity } from "./equipment.js";

test("a drum kit is counted as a set, individual drums are not kits", () => {
  assert.equal(equipmentQuantity({ name: "五鼓架子鼓", quantity: 1 }), "1 套");
  for (const name of ["底鼓", "军鼓", "通鼓", "鼓凳", "吊镲"]) assert.equal(isDrumKit(name), false);
  assert.equal(equipmentQuantity({ name: "吉他音箱", quantity: 2 }), "2 件");
});
