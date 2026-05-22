// Package session 提供 Roundtable 的 SQLite 持久化与事件日志存储。
package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wilenwang/talkaboutit/internal/persona"
	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 数据库操作。
type Store struct {
	db *sql.DB
}

// Roundtable 代表一张圆桌讨论。
type Roundtable struct {
	ID           string     `json:"id"`
	Topic        string     `json:"topic"`
	PersonasJSON string     `json:"personas_json"`
	MaxRounds    int        `json:"max_rounds"`
	Language     string     `json:"language"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	LastEventID  int        `json:"last_event_id"`
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

CREATE TABLE IF NOT EXISTS personas (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS persona_web_profiles (
  persona_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  avatar TEXT NOT NULL,
  role_title TEXT,
  summary TEXT NOT NULL,
  description TEXT,
  archetype TEXT,
  tags_json TEXT NOT NULL DEFAULT '[]',
  sort_order INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS persona_souls (
  persona_id TEXT PRIMARY KEY,
  schema_version TEXT NOT NULL DEFAULT 'soul.v1',
  language_json TEXT NOT NULL DEFAULT '{}',
  identity_json TEXT NOT NULL DEFAULT '{}',
  worldview_json TEXT NOT NULL DEFAULT '{}',
  thinking_json TEXT NOT NULL DEFAULT '{}',
  speaking_style_json TEXT NOT NULL DEFAULT '{}',
  knowledge_json TEXT NOT NULL DEFAULT '{}',
  interaction_json TEXT NOT NULL DEFAULT '{}',
  debate_json TEXT NOT NULL DEFAULT '{}',
  guardrails_json TEXT NOT NULL DEFAULT '{}',
  raw_legacy_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS persona_examples (
  id TEXT PRIMARY KEY,
  persona_id TEXT NOT NULL,
  type TEXT NOT NULL,
  content TEXT NOT NULL,
  language TEXT,
  topic_hint TEXT,
  weight INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS persona_session_states (
  id TEXT PRIMARY KEY,
  roundtable_id TEXT NOT NULL,
  persona_id TEXT NOT NULL,
  used_arguments_json TEXT NOT NULL DEFAULT '[]',
  used_phrases_json TEXT NOT NULL DEFAULT '[]',
  current_focus_json TEXT NOT NULL DEFAULT '{}',
  mood_json TEXT NOT NULL DEFAULT '{}',
  peer_attitudes_json TEXT NOT NULL DEFAULT '{}',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(roundtable_id, persona_id),
  FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE,
  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id) ON DELETE CASCADE
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

// CountPersonas 返回数据库中的 persona 数量。
func (s *Store) CountPersonas(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM personas`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// LoadAll 加载所有 SQLite 中的 Persona，满足 persona.Repository 接口。
func (s *Store) LoadAll() (map[string]persona.Persona, error) {
	rows, err := s.db.Query(
		`SELECT ps.raw_legacy_json
		 FROM personas p
		 JOIN persona_souls ps ON ps.persona_id = p.id
		 WHERE p.status = 'active'
		 ORDER BY p.created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]persona.Persona)
	for rows.Next() {
		var raw sql.NullString
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if !raw.Valid || raw.String == "" {
			continue
		}
		p, err := persona.ValidateJSON([]byte(raw.String))
		if err != nil {
			return nil, err
		}
		result[p.ID] = *p
	}
	return result, rows.Err()
}

// LoadOne 加载指定 Persona，满足 persona.Repository 接口。
func (s *Store) LoadOne(id string) (persona.Persona, error) {
	var raw sql.NullString
	err := s.db.QueryRow(
		`SELECT ps.raw_legacy_json
		 FROM personas p
		 JOIN persona_souls ps ON ps.persona_id = p.id
		 WHERE p.id = ? AND p.status = 'active'`, id,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return persona.Persona{}, fmt.Errorf("persona %s 不存在", id)
	}
	if err != nil {
		return persona.Persona{}, err
	}
	if !raw.Valid || raw.String == "" {
		return persona.Persona{}, fmt.Errorf("persona %s 缺少 legacy payload", id)
	}
	p, err := persona.ValidateJSON([]byte(raw.String))
	if err != nil {
		return persona.Persona{}, err
	}
	if p.ID != id {
		return persona.Persona{}, fmt.Errorf("数据库 ID %q 与 payload ID %q 不匹配", id, p.ID)
	}
	return *p, nil
}

// Save 将 Persona 拆分保存到 web profile、soul 与 examples 表。
func (s *Store) Save(p persona.Persona) error {
	if err := p.Validate(); err != nil {
		return err
	}

	tagsJSON, err := json.Marshal(p.Tags)
	if err != nil {
		return fmt.Errorf("序列化 tags 失败: %w", err)
	}
	rawJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 persona 失败: %w", err)
	}

	languageJSON, _ := json.Marshal(p.Language)
	identityJSON, _ := json.Marshal(map[string]interface{}{
		"name":         p.Name,
		"display_name": p.DisplayName,
		"role_title":   p.RoleTitle,
		"description":  p.Description,
		"tags":         p.Tags,
		"archetype":    p.Archetype,
	})
	worldviewJSON, _ := json.Marshal(map[string]interface{}{
		"stance":       p.Stance,
		"core_beliefs": p.CoreBeliefs,
	})
	thinkingJSON, _ := json.Marshal(map[string]interface{}{})
	speakingJSON, _ := json.Marshal(p.SpeakingStyle)
	knowledgeJSON, _ := json.Marshal(p.KnowledgeScope)
	interactionJSON, _ := json.Marshal(p.InteractionRules)
	debateJSON, _ := json.Marshal(p.DebateGoal)
	guardrailsJSON, _ := json.Marshal(p.Prompting)

	displayName := p.DisplayName
	if displayName == "" {
		displayName = p.Name
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO personas (id, slug, status, updated_at)
		 VALUES (?, ?, 'active', CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET
		   slug = excluded.slug,
		   status = 'active',
		   updated_at = CURRENT_TIMESTAMP`,
		p.ID, p.ID,
	)
	if err != nil {
		return fmt.Errorf("保存 personas 失败: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO persona_web_profiles
		   (persona_id, name, display_name, avatar, role_title, summary, description, archetype, tags_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(persona_id) DO UPDATE SET
		   name = excluded.name,
		   display_name = excluded.display_name,
		   avatar = excluded.avatar,
		   role_title = excluded.role_title,
		   summary = excluded.summary,
		   description = excluded.description,
		   archetype = excluded.archetype,
		   tags_json = excluded.tags_json`,
		p.ID, p.Name, displayName, p.Avatar, p.RoleTitle, p.Description, p.Description, p.Archetype, string(tagsJSON),
	)
	if err != nil {
		return fmt.Errorf("保存 persona_web_profiles 失败: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO persona_souls
		   (persona_id, schema_version, language_json, identity_json, worldview_json, thinking_json,
		    speaking_style_json, knowledge_json, interaction_json, debate_json, guardrails_json, raw_legacy_json, updated_at)
		 VALUES (?, 'soul.v1', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(persona_id) DO UPDATE SET
		   language_json = excluded.language_json,
		   identity_json = excluded.identity_json,
		   worldview_json = excluded.worldview_json,
		   thinking_json = excluded.thinking_json,
		   speaking_style_json = excluded.speaking_style_json,
		   knowledge_json = excluded.knowledge_json,
		   interaction_json = excluded.interaction_json,
		   debate_json = excluded.debate_json,
		   guardrails_json = excluded.guardrails_json,
		   raw_legacy_json = excluded.raw_legacy_json,
		   updated_at = CURRENT_TIMESTAMP`,
		p.ID,
		string(languageJSON),
		string(identityJSON),
		string(worldviewJSON),
		string(thinkingJSON),
		string(speakingJSON),
		string(knowledgeJSON),
		string(interactionJSON),
		string(debateJSON),
		string(guardrailsJSON),
		string(rawJSON),
	)
	if err != nil {
		return fmt.Errorf("保存 persona_souls 失败: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM persona_examples WHERE persona_id = ?`, p.ID); err != nil {
		return fmt.Errorf("清理 persona_examples 失败: %w", err)
	}
	if p.Examples.OpeningLine != "" {
		if _, err := tx.Exec(
			`INSERT INTO persona_examples (id, persona_id, type, content, language, weight)
			 VALUES (?, ?, 'opening', ?, ?, 1)`,
			p.ID+":opening", p.ID, p.Examples.OpeningLine, p.Language.Primary,
		); err != nil {
			return fmt.Errorf("保存 opening example 失败: %w", err)
		}
	}
	if p.Examples.SampleRebuttal != "" {
		if _, err := tx.Exec(
			`INSERT INTO persona_examples (id, persona_id, type, content, language, weight)
			 VALUES (?, ?, 'rebuttal', ?, ?, 1)`,
			p.ID+":rebuttal", p.ID, p.Examples.SampleRebuttal, p.Language.Primary,
		); err != nil {
			return fmt.Errorf("保存 rebuttal example 失败: %w", err)
		}
	}

	return tx.Commit()
}

// Delete 删除指定 Persona，满足 persona.Repository 接口。
func (s *Store) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM personas WHERE id = ?`, id)
	return err
}

// UpsertPersonaSessionState 保存某个 persona 在某场讨论中的运行时状态。
func (s *Store) UpsertPersonaSessionState(ctx context.Context, roundtableID string, personaID string, state *persona.PerPersonaState) error {
	if state == nil {
		state = &persona.PerPersonaState{}
	}
	usedArgumentsJSON, err := json.Marshal(state.UsedArguments)
	if err != nil {
		return fmt.Errorf("序列化 used_arguments 失败: %w", err)
	}
	id := roundtableID + ":" + personaID
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO persona_session_states
		   (id, roundtable_id, persona_id, used_arguments_json, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(roundtable_id, persona_id) DO UPDATE SET
		   used_arguments_json = excluded.used_arguments_json,
		   updated_at = CURRENT_TIMESTAMP`,
		id, roundtableID, personaID, string(usedArgumentsJSON),
	)
	return err
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
