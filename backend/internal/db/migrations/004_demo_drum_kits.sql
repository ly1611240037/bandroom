-- Older demo seed inserted the same B-room kit on every run without a unique constraint.
-- Only remove identical, available, quantity-one rows with that exact legacy description.
DELETE FROM fixed_equipment
WHERE name = '架子鼓' AND description = '房间固定架子鼓'
  AND quantity = 1 AND status = 'available'
  AND room_id IN (SELECT id FROM rooms WHERE name = 'B 房')
  AND id NOT IN (
    SELECT MIN(id) FROM fixed_equipment
    WHERE name = '架子鼓' AND description = '房间固定架子鼓'
      AND quantity = 1 AND status = 'available'
    GROUP BY room_id
  );

UPDATE fixed_equipment
SET description = '一套五鼓配置：底鼓 1、军鼓 1、悬挂通鼓 2、落地通鼓 1；另含踩镲 1 对、吊镲 1、叮叮镲 1、支架与鼓凳，鼓棒自备'
WHERE quantity = 1 AND (
  (name = '架子鼓' AND description IN ('房间固定架子鼓', '含镲片与鼓凳，位于后方节奏区') AND room_id IN (SELECT id FROM rooms WHERE name = 'B 房'))
  OR (name = '五鼓架子鼓' AND description = '含踩镲、吊镲和鼓凳；鼓棒自备' AND room_id IN (SELECT id FROM rooms WHERE name = 'A 房'))
);
