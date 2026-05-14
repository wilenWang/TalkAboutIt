/**
 * Central translation table for TalkAboutIt.
 *
 * Key naming convention:
 *   status.*  — status badges and connection states
 *   action.*  — buttons and CTAs
 *   page.*    — page titles
 *   label.*   — labels and field names
 *   msg.*     — messages, prompts, and descriptions
 *   input.*   — input placeholders
 *   err.*     — error messages
 *   fmt.*     — format patterns (use with f(key, params))
 *   sep.*     — separators and misc
 *
 * To add a new language:
 * 1. Add it to SystemLanguage in LanguageContext.tsx
 * 2. Populate every row
 */

import type { SystemLanguage } from './LanguageContext';

export type TranslationRow = Record<SystemLanguage, string>;

export type TranslationKey = keyof typeof translations;

export const translations = {
  // ── Page titles ──
  pageRoundtable:  { 'zh-CN': '圆桌讨论',           'en-US': 'Roundtable' },
  pageHistory:     { 'zh-CN': '历史记录',           'en-US': 'History' },
  pageReplay:      { 'zh-CN': '回放',              'en-US': 'Replay' },

  // ── Actions / buttons ──
  actionBack:         { 'zh-CN': '返回',              'en-US': 'Back' },
  actionBackToList:   { 'zh-CN': '返回列表',           'en-US': 'Back to list' },
  actionViewReplay:   { 'zh-CN': '查看回放',           'en-US': 'View Replay' },
  actionStart:        { 'zh-CN': '开始讨论',           'en-US': 'Start Discussion' },
  actionRetry:        { 'zh-CN': '重试',              'en-US': 'Retry' },

  // ── Labels ──
  labelParticipants:  { 'zh-CN': '参与者',             'en-US': 'Participants' },
  labelTopic:         { 'zh-CN': '话题',              'en-US': 'Topic' },
  labelLanguage:      { 'zh-CN': '语言',              'en-US': 'Language' },
  labelRounds:        { 'zh-CN': '轮次',              'en-US': 'Rounds' },
  labelDiscussion:    { 'zh-CN': '讨论记录',           'en-US': 'Discussion' },
  labelSearch:        { 'zh-CN': '搜索...',            'en-US': 'Search...' },
  labelTyping:        { 'zh-CN': '正在输入...',         'en-US': 'Typing...' },
  labelLoading:       { 'zh-CN': '加载中...',          'en-US': 'Loading...' },
  labelParticipantsF: { 'zh-CN': '参与者：{names}',       'en-US': 'Participants: {names}' },

  // ── Messages / prompts ──
  msgSelectParticipants:  { 'zh-CN': '请选择 2-4 位参与者',   'en-US': 'Select 2-4 participants' },
  msgMinParticipants:     { 'zh-CN': '请至少选择 2 位参与者',   'en-US': 'Select at least 2 participants' },
  msgDiscussionSubtitle:  { 'zh-CN': '选择参与者，设定话题，观察 AI 人格之间的对话。',
                             'en-US': 'Choose participants, set a topic, and observe the conversation between AI personas.' },
  msgHistorySubtitle:     { 'zh-CN': '回顾已完成的圆桌讨论',    'en-US': 'Review completed roundtable discussions' },
  msgStartDiscussionCta:  { 'zh-CN': '去开始一场讨论吧',      'en-US': 'Start a discussion' },
  msgNoMessages:          { 'zh-CN': '暂无消息',           'en-US': 'No messages yet' },
  msgNotStarted:          { 'zh-CN': '讨论尚未开始',         'en-US': 'Discussion has not started yet' },
  msgNoMessagesInDiscussion: { 'zh-CN': '该讨论暂无消息记录', 'en-US': 'No messages in this discussion' },
  msgNoHistory:           { 'zh-CN': '暂无历史记录',         'en-US': 'No history yet' },
  msgPreparing:           { 'zh-CN': '准备中...',          'en-US': 'Preparing...' },
  msgLoading:             { 'zh-CN': '加载中...',          'en-US': 'Loading...' },

  // ── Status / connection ──
  statusPending:        { 'zh-CN': '待开始',             'en-US': 'Pending' },
  statusRunning:        { 'zh-CN': '进行中',             'en-US': 'Running' },
  statusCompleted:      { 'zh-CN': '已完成',             'en-US': 'Completed' },
  statusFailed:         { 'zh-CN': '失败',              'en-US': 'Failed' },
  statusInProgress:     { 'zh-CN': '讨论进行中',          'en-US': 'Discussion in progress' },
  statusConnecting:     { 'zh-CN': '连接中...',          'en-US': 'Connecting...' },
  statusReconnecting:   { 'zh-CN': '重连中...',          'en-US': 'Reconnecting...' },
  statusDisconnected:   { 'zh-CN': '已断开',             'en-US': 'Disconnected' },
  statusConnectionLost: { 'zh-CN': '连接已断开，正在重连...', 'en-US': 'Connection lost, reconnecting...' },
  statusReconnectManually: { 'zh-CN': '手动重连',        'en-US': 'Reconnect manually' },

  // ── Errors ──
  errUnknown:             { 'zh-CN': '未知错误',           'en-US': 'Unknown error' },
  errLoadFailed:          { 'zh-CN': '加载失败',           'en-US': 'Failed to load' },
  errSnapshotFailed:      { 'zh-CN': '加载快照失败',        'en-US': 'Failed to load snapshot' },
  errReplayFailed:        { 'zh-CN': '加载回放失败',        'en-US': 'Failed to load replay' },
  errFetchPersonas:       { 'zh-CN': '加载参与者失败',      'en-US': 'Failed to load participants' },
  errCreateDiscussion:    { 'zh-CN': '创建讨论失败',        'en-US': 'Failed to create discussion' },
  errStartDiscussion:     { 'zh-CN': '启动讨论失败',        'en-US': 'Failed to start discussion' },
  errListDiscussions:     { 'zh-CN': '获取讨论列表失败',     'en-US': 'Failed to load discussions' },
  errGetSnapshot:         { 'zh-CN': '获取讨论快照失败',     'en-US': 'Failed to load discussion snapshot' },

  // ── Input placeholders ──
  inputTopicPlaceholder:  { 'zh-CN': '输入讨论话题...',      'en-US': 'Enter a discussion topic...' },

  // ── Separators ──
  sepParticipants:        { 'zh-CN': '、',               'en-US': ', ' },

  // ── Format patterns (use with f(key, params)) ──
  fmtSelectedCount:    { 'zh-CN': '已选择 {n} 人',        'en-US': '{n} selected' },
  fmtMessageCount:     { 'zh-CN': '{n} 条消息',           'en-US': '{n} messages' },
  fmtRoundLabel:       { 'zh-CN': '第 {n} 轮',            'en-US': 'Round {n}' },
  fmtRoundCount:       { 'zh-CN': '{n} 轮',               'en-US': '{n} rounds' },
  fmtStatusLabel:      { 'zh-CN': '状态：{s}',             'en-US': 'Status: {s}' },
  fmtSummaryLine:      { 'zh-CN': '共 {msg} 条消息 · {rnd} 轮', 'en-US': '{msg} messages · {rnd} rounds' },
} as const;

/** Maps each format key to its expected parameter types. */
export interface FormatParams {
  fmtSelectedCount:    { n: number };
  fmtMessageCount:     { n: number };
  fmtRoundLabel:       { n: number };
  fmtRoundCount:       { n: number };
  fmtStatusLabel:      { s: string };
  fmtParticipantsLabel: { names: string };
  fmtSummaryLine:      { msg: number; rnd: number };
}
