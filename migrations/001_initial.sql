PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS boards (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS lists (
    id         TEXT PRIMARY KEY,
    board_id   TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_lists_board_id ON lists(board_id);

CREATE TABLE IF NOT EXISTS cards (
    id          TEXT PRIMARY KEY,
    list_id     TEXT NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL DEFAULT 0,
    assignee    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'unassigned' CHECK(status IN ('unassigned','assigned','in_progress','blocked','done')),
    priority    TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low','medium','high','critical')),
    due_date    TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_cards_list_id ON cards(list_id);

CREATE TABLE IF NOT EXISTS card_dependencies (
    id                 TEXT PRIMARY KEY,
    card_id            TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    depends_on_card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE(card_id, depends_on_card_id)
);
CREATE INDEX IF NOT EXISTS idx_card_deps_card_id ON card_dependencies(card_id);
CREATE INDEX IF NOT EXISTS idx_card_deps_depends_on ON card_dependencies(depends_on_card_id);

CREATE TABLE IF NOT EXISTS labels (
    id       TEXT PRIMARY KEY,
    board_id TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name     TEXT NOT NULL,
    color    TEXT NOT NULL DEFAULT '#6b7280'
);
CREATE INDEX IF NOT EXISTS idx_labels_board_id ON labels(board_id);

CREATE TABLE IF NOT EXISTS card_labels (
    card_id  TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    label_id TEXT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, label_id)
);

CREATE TABLE IF NOT EXISTS activity_log (
    id         TEXT PRIMARY KEY,
    card_id    TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    actor      TEXT NOT NULL,
    action     TEXT NOT NULL CHECK(action IN ('created','moved','assigned','unassigned','status_changed','comment','dependency_added','dependency_removed','label_added','label_removed')),
    detail     TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_activity_card_id ON activity_log(card_id);
