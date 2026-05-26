import { useLanguage } from '../i18n/LanguageContext';
import type { SSEConnectionStatus } from '../hooks/useSSE';
import type { AppStatus } from '../stores/useAppStore';

interface Props {
  sseStatus: SSEConnectionStatus;
  appStatus: AppStatus;
}

export default function ConnectionBadge({ sseStatus, appStatus }: Props) {
  const { t } = useLanguage();

  if (sseStatus === 'reconnecting') {
    return <span className="text-xs text-orange-500 font-medium animate-pulse">↻ {t('statusReconnecting')}</span>;
  }
  if (sseStatus === 'connecting') {
    return <span className="text-xs text-[#0075de] font-medium animate-pulse">● {t('statusConnecting')}</span>;
  }
  if (sseStatus === 'disconnected') {
    return <span className="text-xs text-red-500 font-medium">✗ {t('statusDisconnected')}</span>;
  }
  if (appStatus === 'streaming') {
    return <span className="text-xs text-[#0075de] font-medium animate-pulse">● {t('statusInProgress')}</span>;
  }
  if (appStatus === 'completed') {
    return <span className="text-xs text-green-600 font-medium">✓ {t('statusCompleted')}</span>;
  }
  return null;
}
