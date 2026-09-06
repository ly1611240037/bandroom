CREATE TABLE IF NOT EXISTS venue_settings (
    setting_key TEXT PRIMARY KEY,
    setting_value TEXT NOT NULL
);

INSERT OR IGNORE INTO venue_settings(setting_key, setting_value) VALUES ('cleaning_buffer_minutes', '30');

INSERT OR IGNORE INTO business_hours(weekday, opens_at, closes_at, enabled) VALUES
    (0, '08:00', '23:00', 1), (1, '08:00', '23:00', 1), (2, '08:00', '23:00', 1),
    (3, '08:00', '23:00', 1), (4, '08:00', '23:00', 1), (5, '08:00', '23:00', 1),
    (6, '08:00', '23:00', 1);
