import { useEffect, useRef, useCallback, useState } from 'react';

export interface SSEMessage {
  id: string;
  event: string;
  data: Record<string, unknown>;
}

export type SSEConnectionStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected';

export interface UseSSEOptions {
  /** 初始 Last-Event-ID，用于首次连接时从指定位置恢复 */
  initialLastEventId?: string;
}

export function useSSE(
  url: string | null,
  onMessage: (msg: SSEMessage) => void,
  onError?: (err: Event) => void,
  options?: UseSSEOptions
) {
  const esRef = useRef<EventSource | null>(null);
  const lastEventIdRef = useRef<string>(options?.initialLastEventId ?? '');
  const retryCountRef = useRef(0);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // 组件卸载标记，防止卸载后仍执行重连或 setStatus
  const disposedRef = useRef(false);
  // 连接序号，用于区分新旧连接，旧连接的 onerror 直接忽略
  const connectionSeqRef = useRef(0);
  // 将回调放入 ref，避免 connect effect 依赖变化导致反复重建
  const onMessageRef = useRef(onMessage);
  const onErrorRef = useRef(onError);

  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    onErrorRef.current = onError;
  }, [onError]);

  const [status, setStatus] = useState<SSEConnectionStatus>('idle');

  // 安全设置状态：若已卸载则忽略
  const setStatusSafe = useCallback((s: SSEConnectionStatus) => {
    if (!disposedRef.current) {
      setStatus(s);
    }
  }, []);

  // 清理重连定时器
  const clearRetryTimer = useCallback(() => {
    if (retryTimerRef.current) {
      clearTimeout(retryTimerRef.current);
      retryTimerRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    if (!url) return;
    clearRetryTimer();

    // 关闭旧连接
    if (esRef.current) {
      esRef.current.close();
    }

    const isReconnect = retryCountRef.current > 0;
    setStatusSafe(isReconnect ? 'reconnecting' : 'connecting');

    // 首次或重连都带上 Last-Event-ID（如果有）
    const connectUrl = lastEventIdRef.current
      ? `${url}${url.includes('?') ? '&' : '?'}lastEventId=${encodeURIComponent(lastEventIdRef.current)}`
      : url;

    const es = new EventSource(connectUrl);
    esRef.current = es;

    // 分配新连接序号
    const seq = ++connectionSeqRef.current;

    es.onopen = () => {
      if (disposedRef.current || seq !== connectionSeqRef.current) return;
      retryCountRef.current = 0;
      setStatusSafe('connected');
    };

    es.onmessage = (e) => {
      if (e.lastEventId) {
        lastEventIdRef.current = e.lastEventId;
      }
      try {
        const data = JSON.parse(e.data);
        onMessageRef.current({
          id: e.lastEventId,
          event: e.type,
          data,
        });
      } catch {
        onMessageRef.current({
          id: e.lastEventId,
          event: e.type,
          data: { raw: e.data },
        });
      }
    };

    // 处理命名事件
    const eventTypes = [
      'stream_start',
      'round_start',
      'speaking',
      'message_chunk',
      'message_done',
      'message_aborted',
      'round_end',
      'stream_done',
      'error',
    ];

    eventTypes.forEach((type) => {
      es.addEventListener(type, (e) => {
        const me = e as MessageEvent;
        if (me.lastEventId) {
          lastEventIdRef.current = me.lastEventId;
        }
        try {
          const data = JSON.parse(me.data);
          onMessageRef.current({
            id: me.lastEventId,
            event: type,
            data,
          });
        } catch {
          onMessageRef.current({
            id: me.lastEventId,
            event: type,
            data: { raw: me.data },
          });
        }
      });
    });

    es.onerror = (err) => {
      // 若组件已卸载或该连接已过期，直接忽略
      if (disposedRef.current || seq !== connectionSeqRef.current) return;

      if (onErrorRef.current) onErrorRef.current(err);
      es.close();
      if (esRef.current === es) {
        esRef.current = null;
      }

      // 指数退避重连：1s / 2s / 4s / 8s，最多 5 次
      const maxRetries = 5;
      if (retryCountRef.current < maxRetries) {
        const delay = Math.min(1000 * Math.pow(2, retryCountRef.current), 8000);
        retryCountRef.current++;
        setStatusSafe('reconnecting');
        retryTimerRef.current = setTimeout(() => {
          // 再次检查 disposed 与序号
          if (!disposedRef.current && seq === connectionSeqRef.current) {
            connect();
          }
        }, delay);
      } else {
        setStatusSafe('disconnected');
      }
    };
  }, [url, clearRetryTimer, setStatusSafe]);

  useEffect(() => {
    // 每次 url 变化时重置 lastEventId（除非 options 提供了初始值）
    lastEventIdRef.current = options?.initialLastEventId ?? '';
    retryCountRef.current = 0;
    disposedRef.current = false;
    connect();
    return () => {
      disposedRef.current = true;
      clearRetryTimer();
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
    };
  }, [url, connect, clearRetryTimer, options?.initialLastEventId]);

  return { reconnect: connect, status };
}
