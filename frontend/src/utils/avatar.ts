/**
 * Check if an avatar string represents an image URL.
 * Returns true for http(s) URLs and paths containing '/'.
 */
export function isImageAvatar(avatar: string): boolean {
  return avatar.startsWith('http') || avatar.includes('/');
}
