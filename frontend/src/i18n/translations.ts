/**
 * Central translation table for TalkAboutIt.
 *
 * To add a new language:
 * 1. Add a column to TranslationTable
 * 2. Populate every row
 * 3. Add the lang to the Language type in LanguageContext
 */

export interface TranslationRow {
  zh: string;
  en: string;
}

export type TranslationKey = keyof typeof translations;

export const translations = {
  // ── Header ──
  roundtable:         { zh: '圆桌讨论',           en: 'Roundtable' },
  history:            { zh: '历史记录',           en: 'History' },
  back:               { zh: '返回',              en: 'Back' },
  backToList:         { zh: '返回列表',           en: 'Back to list' },
  viewReplay:         { zh: '查看回放',           en: 'View Replay' },
  replay:             { zh: '回放',              en: 'Replay' },

  // ── Sidebar / PersonaSelector ──
  participants:       { zh: '参与者',             en: 'Participants' },
  search:             { zh: '搜索...',            en: 'Search...' },
  selectParticipants: { zh: '请选择 2-4 位参与者',   en: 'Select 2-4 participants' },

  // ── Controls ──
  topic:              { zh: '话题',              en: 'Topic' },
  topicPlaceholder:   { zh: '输入讨论话题...',      en: 'Enter a discussion topic...' },
  language:           { zh: '语言',              en: 'Language' },
  rounds:             { zh: '轮次',              en: 'Rounds' },
  startDiscussion:    { zh: '开始讨论',           en: 'Start Discussion' },
  preparing:          { zh: '准备中...',          en: 'Preparing...' },
  minParticipants:    { zh: '请至少选择 2 位参与者',   en: 'Select at least 2 participants' },

  // ── Discussion area ──
  discussion:         { zh: '讨论记录',           en: 'Discussion' },
  noMessages:         { zh: '暂无消息',           en: 'No messages yet' },
  discussionSubtitle: { zh: '选择参与者，设定话题，观察 AI 人格之间的对话。',
                         en: 'Choose participants, set a topic, and observe the conversation between AI personas.' },
  discussionReplay:   { zh: '讨论回放',           en: 'Replay' },
  notStarted:         { zh: '讨论尚未开始',         en: 'Discussion has not started yet' },
  noMessagesInDiscussion: { zh: '该讨论暂无消息记录', en: 'No messages in this discussion' },

  // ── Connection status ──
  reconnecting:       { zh: '重连中...',          en: 'Reconnecting...' },
  connecting:         { zh: '连接中...',          en: 'Connecting...' },
  disconnected:       { zh: '已断开',             en: 'Disconnected' },
  inProgress:         { zh: '讨论进行中',          en: 'Discussion in progress' },
  reconnectManually:  { zh: '手动重连',           en: 'Reconnect manually' },
  connectionLost:     { zh: '连接已断开，正在重连...', en: 'Connection lost, reconnecting...' },

  // ── Status badges ──
  statusPending:      { zh: '待开始',             en: 'Pending' },
  statusRunning:      { zh: '进行中',             en: 'Running' },
  statusCompleted:    { zh: '已完成',             en: 'Completed' },
  statusFailed:       { zh: '失败',              en: 'Failed' },

  // ── History ──
  participantsLabel:  { zh: '参与者：',           en: 'Participants: ' },
  historySubtitle:    { zh: '回顾已完成的圆桌讨论',    en: 'Review completed roundtable discussions' },
  noHistory:          { zh: '暂无历史记录',         en: 'No history yet' },
  startDiscussionCta: { zh: '去开始一场讨论吧',      en: 'Start a discussion' },
  loading:            { zh: '加载中...',          en: 'Loading...' },

  // ── Errors ──
  retry:              { zh: '重试',              en: 'Retry' },
  unknownError:       { zh: '未知错误',           en: 'Unknown error' },
  loadFailed:         { zh: '加载失败',           en: 'Failed to load' },
  snapshotFailed:     { zh: '加载快照失败',        en: 'Failed to load snapshot' },
  replayFailed:       { zh: '加载回放失败',        en: 'Failed to load replay' },
  fetchPersonasFailed: { zh: '加载参与者失败',      en: 'Failed to load participants' },
  createDiscussionFailed: { zh: '创建讨论失败',    en: 'Failed to create discussion' },
  startDiscussionFailed:  { zh: '启动讨论失败',    en: 'Failed to start discussion' },
  listDiscussionsFailed:  { zh: '获取讨论列表失败', en: 'Failed to load discussions' },
  getSnapshotFailed:      { zh: '获取讨论快照失败', en: 'Failed to load discussion snapshot' },

  // ── Misc ──
  typing:             { zh: '正在输入...',         en: 'Typing...' },
  participantsSeparator: { zh: '、',             en: ', ' },

  // ── Dynamic patterns (use with f(key, params)) ──
  selectedCount:      { zh: '已选择 {n} 人',        en: '{n} selected' },
  messageCount:       { zh: '{n} 条消息',           en: '{n} messages' },
  roundLabel:         { zh: '第 {n} 轮',               en: 'Round {n}' },
  roundCount:         { zh: '{n} 轮',               en: '{n} rounds' },
  statusLabel:        { zh: '状态：{s}',             en: 'Status: {s}' },
  participantsLabelF: { zh: '参与者：{names}',       en: 'Participants: {names}' },
  summaryLine:        { zh: '共 {msg} 条消息 · {rnd} 轮', en: '{msg} messages · {rnd} rounds' },
} as const;
