import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import TopicPanel from './TopicPanel';
import { LanguageProvider } from '../i18n/LanguageContext';

function renderTopicPanel(onStart = vi.fn()) {
  const onTopicChange = vi.fn();
  render(
    <LanguageProvider>
      <TopicPanel
        topic="Base topic"
        onTopicChange={onTopicChange}
        rounds={3}
        onRoundsChange={vi.fn()}
        onStart={onStart}
        canStart={true}
        loading={false}
      />
    </LanguageProvider>
  );

  return { onStart, onTopicChange };
}

describe('TopicPanel', () => {
  it('passes the merged file contents to start instead of relying on async state', async () => {
    const { onStart, onTopicChange } = renderTopicPanel();
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['Evidence from upload'], 'notes.txt', { type: 'text/plain' });

    fireEvent.change(input, { target: { files: [file] } });
    fireEvent.click(screen.getByRole('button', { name: /talk/i }));

    await waitFor(() => {
      expect(onStart).toHaveBeenCalledWith(expect.stringContaining('Evidence from upload'));
    });
    expect(onTopicChange).toHaveBeenCalledWith(expect.stringContaining('--- notes.txt ---'));
  });
});
