// Package session 提供 Roundtable 的 SQLite 持久化与事件日志存储。
package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 数据库操作。
type Store struct {
	db *sql.DB
}

// Roundtable 代表一张圆桌讨论。
type Roundtable struct {
	ID           string    `json:"id"`
	Topic        string    `json:"topic"`
	PersonasJSON string    `json:"personas_json"`
	MaxRounds    int       `json:"max_rounds"`
	Language     string    `json:"language"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	LastEventID  int       `json:"last_event_id"`
}

// Message 代表一条发言消息。
type Message struct {
	ID           string    `json:"id"`
	RoundtableID string    `json:"roundtable_id"`
	Round        int       `json:"round"`
	SpeakerIndex int       `json:"speaker_index"`
	PersonaID    string    `json:"persona_id"`
	Content      string    `json:"content"`
	EventID      int       `json:"event_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// Event 代表一个 SSE 事件日志条目。
type Event struct {
	RoundtableID string    `json:"roundtable_id"`
	EventID      int       `json:"event_id"`
	EventType    string    `json:"event_type"`
	Round        *int      `json:"round,omitempty"`
	SpeakerIndex *int      `json:"speaker_index,omitempty"`
	PersonaID    *string   `json:"persona_id,omitempty"`
	MessageID    *string   `json:"message_id,omitempty"`
	PayloadJSON  string    `json:"payload_json"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewStore 打开（或创建）SQLite 数据库并初始化表结构。
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	store := &Store{db: db}
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}
	return store, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS roundtables (
  id TEXT PRIMARY KEY,
  topic TEXT NOT NULL,
  personas_json TEXT NOT NULL,
  max_rounds INTEGER NOT NULL DEFAULT 3,
  language TEXT NOT NULL DEFAULT 'zh-CN',
  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at DATETIME,
  finished_at DATETIME,
  last_event_id INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS messages (
  id TEXT PRIMARY KEY,
  roundtable_id TEXT NOT NULL,
  round INTEGER NOT NULL,
  speaker_index INTEGER NOT NULL,
  persona_id TEXT NOT NULL,
  content TEXT NOT NULL,
  event_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id),
  UNIQUE(roundtable_id, round, speaker_index),
  UNIQUE(roundtable_id, event_id)
);

CREATE TABLE IF NOT EXISTS roundtable_events (
  roundtable_id TEXT NOT NULL,
  event_id INTEGER NOT NULL,
  event_type TEXT NOT NULL CHECK (
    event_type IN (
      'stream_start', 'round_start', 'speaking', 'message_chunk',
      'message_done', 'message_aborted', 'round_end', 'stream_done', 'error'
    )
  ),
  round INTEGER,
  speaker_index INTEGER,
  persona_id TEXT,
  message_id TEXT,
  payload_json TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (roundtable_id, event_id),
  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id)
);
`
	_, err := s.db.Exec(schema)
	return err
}

// CreateRoundtable 创建一张新的圆桌讨论。
func (s *Store) CreateRoundtable(ctx context.Context, rt *Roundtable) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO roundtables (id, topic, personas_json, max_rounds, language, status, last_event_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rt.ID, rt.Topic, rt.PersonasJSON, rt.MaxRounds, rt.Language, rt.Status, rt.LastEventID,
	)
	return err
}

// GetRoundtable 获取指定 ID 的圆桌讨论。
func (s *Store) GetRoundtable(ctx context.Context, id string) (*Roundtable, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, topic, personas_json, max_rounds, language, status, created_at, started_at, finished_at, last_event_id
		 FROM roundtables WHERE id = ?`, id)

	rt := &Roundtable{}
	var startedAt, finishedAt sql.NullTime
	err := row.Scan(&rt.ID, &rt.Topic, &rt.PersonasJSON, &rt.MaxRounds, &rt.Language, &rt.Status,
		&rt.CreatedAt, &startedAt, &finishedAt, &rt.LastEventID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("roundtable %s 不存在", id)
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		rt.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		rt.FinishedAt = &finishedAt.Time
	}
	return rt, nil
}

// UpdateStatus 更新圆桌讨论的状态。
func (s *Store) UpdateStatus(ctx context.Context, id string, status string) error {
	var startedAt, finishedAt interface{}
	now := time.Now().UTC()

	switch status {
	case "running":
		startedAt = now
	case "completed", "failed":
		finishedAt = now
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE roundtables SET status = ?, started_at = COALESCE(?, started_at), finished_at = COALESCE(?, finished_at) WHERE id = ?`,
		status, startedAt, finishedAt, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("roundtable %s 不存在或状态未变更", id)
	}
	return nil
}

// MarkRunning 原子地将 roundtable 从 pending 切换到 running。
// 返回 true 表示切换成功，false 表示未切换（可能已被其他请求启动）。
func (s *Store) MarkRunning(ctx context.Context, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE roundtables SET status = 'running', started_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'pending'`,
		id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// AddEvent 在事务中写入事件日志并更新 last_event_id。
// 若 eventType 为 message_done，则同时写入 messages 表。
func (s *Store) AddEvent(ctx context.Context, roundtableID string, eventType string,
	round, speakerIndex *int, personaID, messageID *string, payload map[string]interface{}) (*Event, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 读取当前 last_event_id
	var lastEventID int
	err = tx.QueryRowContext(ctx, `SELECT last_event_id FROM roundtables WHERE id = ?`, roundtableID).Scan(&lastEventID)
	if err != nil {
		return nil, fmt.Errorf("读取 last_event_id 失败: %w", err)
	}

	nextEventID := lastEventID + 1

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化 payload 失败: %w", err)
	}

	// 插入事件
	_, err = tx.ExecContext(ctx,
		`INSERT INTO roundtable_events (roundtable_id, event_id, event_type, round, speaker_index, persona_id, message_id, payload_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		roundtableID, nextEventID, eventType,
		nullInt(round), nullInt(speakerIndex), nullStr(personaID), nullStr(messageID),
		string(payloadJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("插入事件失败: %w", err)
	}

	// 更新 last_event_id
	_, err = tx.ExecContext(ctx, `UPDATE roundtables SET last_event_id = ? WHERE id = ?`, nextEventID, roundtableID)
	if err != nil {
		return nil, fmt.Errorf("更新 last_event_id 失败: %w", err)
	}

	// 若是 message_done，写入 messages 表（upsert 保证幂等）
	if eventType == "message_done" && messageID != nil && personaID != nil && round != nil && speakerIndex != nil {
		content, _ := payload["content"].(string)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO messages (id, roundtable_id, round, speaker_index, persona_id, content, event_id)
				 VALUES (?, ?, ?, ?, ?, ?, ?)
				 ON CONFLICT(id) DO UPDATE SET
				   round = excluded.round,
				   speaker_index = excluded.speaker_index,
				   persona_id = excluded.persona_id,
				   content = excluded.content,
				   event_id = excluded.event_id`,
			*messageID, roundtableID, *round, *speakerIndex, *personaID, content, nextEventID,
		)
		if err != nil {
			return nil, fmt.Errorf("插入消息失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &Event{
		RoundtableID: roundtableID,
		EventID:      nextEventID,
		EventType:    eventType,
		Round:        round,
		SpeakerIndex: speakerIndex,
		PersonaID:    personaID,
		MessageID:    messageID,
		PayloadJSON:  string(payloadJSON),
		CreatedAt:    time.Now().UTC(),
	}, nil
}

// AddMessage 直接写入 messages 表（用于非事件驱动场景）。
func (s *Store) AddMessage(ctx context.Context, msg *Message) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (id, roundtable_id, round, speaker_index, persona_id, content, event_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.RoundtableID, msg.Round, msg.SpeakerIndex, msg.PersonaID, msg.Content, msg.EventID,
	)
	return err
}

// GetMessages 获取指定圆桌的所有消息。
func (s *Store) GetMessages(ctx context.Context, roundtableID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, roundtable_id, round, speaker_index, persona_id, content, event_id, created_at
		 FROM messages WHERE roundtable_id = ? ORDER BY round, speaker_index`,
		roundtableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.RoundtableID, &m.Round, &m.SpeakerIndex, &m.PersonaID, &m.Content, &m.EventID, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// ListRoundtables 按条件查询圆桌讨论列表，按 created_at DESC 排序。
func (s *Store) ListRoundtables(ctx context.Context, status string, limit int) ([]Roundtable, error) {
	if limit <= 0 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, topic, personas_json, max_rounds, language, status, created_at, started_at, finished_at, last_event_id
			 FROM roundtables WHERE status = ? ORDER BY created_at DESC LIMIT ?`,
			status, limit)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, topic, personas_json, max_rounds, language, status, created_at, started_at, finished_at, last_event_id
			 FROM roundtables ORDER BY created_at DESC LIMIT ?`,
			limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Roundtable
	for rows.Next() {
		var rt Roundtable
		var startedAt, finishedAt sql.NullTime
		err := rows.Scan(&rt.ID, &rt.Topic, &rt.PersonasJSON, &rt.MaxRounds, &rt.Language, &rt.Status,
			&rt.CreatedAt, &startedAt, &finishedAt, &rt.LastEventID)
		if err != nil {
			return nil, err
		}
		if startedAt.Valid {
			rt.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			rt.FinishedAt = &finishedAt.Time
		}
		list = append(list, rt)
	}
	return list, rows.Err()
}

// GetEventsAfter 获取指定 event_id 之后的所有事件（用于 SSE 重连补发）。
func (s *Store) GetEventsAfter(ctx context.Context, roundtableID string, afterEventID int) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT roundtable_id, event_id, event_type, round, speaker_index, persona_id, message_id, payload_json, created_at
		 FROM roundtable_events WHERE roundtable_id = ? AND event_id > ? ORDER BY event_id`,
		roundtableID, afterEventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var round, speakerIndex sql.NullInt64
		var personaID, messageID sql.NullString
		err := rows.Scan(&e.RoundtableID, &e.EventID, &e.EventType, &round, &speakerIndex, &personaID, &messageID, &e.PayloadJSON, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		if round.Valid {
			r := int(round.Int64)
			e.Round = &r
		}
		if speakerIndex.Valid {
			si := int(speakerIndex.Int64)
			e.SpeakerIndex = &si
		}
		if personaID.Valid {
			e.PersonaID = &personaID.String
		}
		if messageID.Valid {
			e.MessageID = &messageID.String
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func nullInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullStr(v *string) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
