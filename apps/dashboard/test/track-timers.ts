const originalSetTimeout = globalThis.setTimeout;
const activeTimeouts = new Set<unknown>();

globalThis.setTimeout = function(
  cb: Parameters<typeof originalSetTimeout>[0],
  ms?: number,
  ...args: unknown[]
) {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore: spread arguments type mismatch is expected here
  const id = originalSetTimeout(cb, ms, ...args);
  activeTimeouts.add(id);
  const stack = new Error().stack;
  console.log('SETTIMEOUT', id, ms, stack?.split('\n')[2]);
  return id;
} as typeof originalSetTimeout;

const originalClearTimeout = globalThis.clearTimeout;

globalThis.clearTimeout = function(id: Parameters<typeof originalClearTimeout>[0]) {
  activeTimeouts.delete(id);
  return originalClearTimeout(id);
} as typeof originalClearTimeout;
