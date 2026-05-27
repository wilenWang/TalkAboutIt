import { describe, expect, it } from 'vitest';
import { translations } from './translations';

describe('translations', () => {
  it('defines the participants format row used by replay pages', () => {
    expect(translations.fmtParticipantsLabel['zh-CN']).toBe('参与者：{names}');
    expect(translations.fmtParticipantsLabel['en-US']).toBe('Participants: {names}');
  });
});
