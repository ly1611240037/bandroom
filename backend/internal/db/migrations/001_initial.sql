CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role TEXT NOT NULL CHECK (role IN ('owner', 'customer')),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    email_verified_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS membership_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_type TEXT NOT NULL CHECK (plan_type IN ('monthly', 'count')),
    name TEXT NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    included_uses INTEGER,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plan_type, name)
);

CREATE TABLE IF NOT EXISTS membership_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    plan_id INTEGER NOT NULL REFERENCES membership_plans(id),
    starts_at TEXT NOT NULL,
    ends_at TEXT,
    remaining_uses INTEGER,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'active', 'expired', 'depleted', 'disabled')),
    paid_amount_cents INTEGER NOT NULL DEFAULT 0 CHECK (paid_amount_cents >= 0),
    payment_method TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    capacity INTEGER,
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'maintenance', 'suspended')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS room_photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS fixed_equipment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 0),
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'maintenance', 'disabled')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public_equipment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    hourly_price_cents INTEGER NOT NULL DEFAULT 0 CHECK (hourly_price_cents >= 0),
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'maintenance', 'disabled')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS business_hours (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekday INTEGER NOT NULL UNIQUE CHECK (weekday BETWEEN 0 AND 6),
    opens_at TEXT NOT NULL,
    closes_at TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1))
);

CREATE TABLE IF NOT EXISTS closures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    starts_at TEXT NOT NULL,
    ends_at TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    room_id INTEGER NOT NULL REFERENCES rooms(id),
    membership_card_id INTEGER NOT NULL REFERENCES membership_cards(id),
    band_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    starts_at TEXT NOT NULL,
    ends_at TEXT NOT NULL,
    occupied_until TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show')),
    notes TEXT NOT NULL DEFAULT '',
    cancellation_reason TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS booking_equipment (
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    equipment_id INTEGER NOT NULL REFERENCES public_equipment(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (booking_id, equipment_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    read_at TEXT,
    email_status TEXT NOT NULL DEFAULT 'pending' CHECK (email_status IN ('pending', 'sent', 'failed', 'not_required')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS issue_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    room_id INTEGER REFERENCES rooms(id),
    equipment_id INTEGER REFERENCES public_equipment(id),
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TEXT
);

CREATE TABLE IF NOT EXISTS venue_content (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_key TEXT NOT NULL UNIQUE,
    content_value TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_user_id INTEGER NOT NULL REFERENCES users(id),
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER,
    before_json TEXT,
    after_json TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bookings_room_interval ON bookings(room_id, starts_at, occupied_until);
CREATE INDEX IF NOT EXISTS idx_bookings_user_status ON bookings(user_id, status);
CREATE INDEX IF NOT EXISTS idx_membership_cards_user_dates ON membership_cards(user_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read_at);
