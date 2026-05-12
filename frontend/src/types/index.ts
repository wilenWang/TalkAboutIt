export interface PersonaSummary {
  id: string;
  name: string;
  display_name: string;
  avatar: string;
  role_title: string;
  description: string;
  tags: string[];
}

export interface Roundtable {
  id: string;
  topic: string;
  personas: string[];
  max_rounds: number;
  language: string;
  status: string;
  created_at: string;
}

export interface SSEEvent {
  id: string;
  event: string;
  data: Record<string, unknown>;
}

export interface Message {
  id: string;
  roundtable_id: string;
  round: number;
  speaker_index: number;
  persona_id: string;
  content: string;
  event_id: number;
  created_at: string;
}

// 回放页使用的消息类型
export interface ReplayMessage {
  id: string;
  avatar: string;
  author: string;
  personaId: string;
  round: number;
  content: string;
}
