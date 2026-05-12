const BASE = '';

export interface RoundtableListItem {
  id: string;
  topic: string;
  personas: string[];
  max_rounds: number;
  status: string;
  created_at: string;
  finished_at?: string;
}

export async function fetchPersonas(): Promise<import('../types').PersonaSummary[]> {
  const res = await fetch(`${BASE}/api/v1/personas`);
  if (!res.ok) throw new Error('加载 persona 失败');
  return res.json();
}

export async function createRoundtable(body: {
  topic: string;
  personas: string[];
  max_rounds: number;
  language: string;
}): Promise<import('../types').Roundtable> {
  const res = await fetch(`${BASE}/api/v1/roundtables`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error('创建讨论失败');
  return res.json();
}

export async function startRoundtable(id: string): Promise<{ id: string; status: string }> {
  const res = await fetch(`${BASE}/api/v1/roundtables/${id}/start`, { method: 'POST' });
  if (!res.ok) throw new Error('启动讨论失败');
  return res.json();
}

// 获取 roundtable 列表
export async function listRoundtables(status?: string): Promise<RoundtableListItem[]> {
  const qs = status ? `?status=${encodeURIComponent(status)}` : '';
  const res = await fetch(`${BASE}/api/v1/roundtables${qs}`);
  if (!res.ok) throw new Error('获取讨论列表失败');
  return res.json();
}

// 获取 roundtable 快照（含历史消息）
export async function getRoundtable(id: string): Promise<{
  id: string;
  topic: string;
  personas: string[];
  max_rounds: number;
  language: string;
  status: string;
  created_at: string;
  started_at?: string;
  finished_at?: string;
  last_event_id: number;
  messages: {
    id: string;
    roundtable_id: string;
    round: number;
    speaker_index: number;
    persona_id: string;
    content: string;
    event_id: number;
    created_at: string;
  }[];
}> {
  const res = await fetch(`${BASE}/api/v1/roundtables/${id}`);
  if (!res.ok) throw new Error('获取讨论快照失败');
  return res.json();
}
