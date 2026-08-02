
const originalSetTimeout = globalThis.setTimeout;
const activeTimeouts = new Set();
globalThis.setTimeout = function(cb, ms, ...args) {
  const id = originalSetTimeout(cb, ms, ...args);
  activeTimeouts.add(id);
  const stack = new Error().stack;
  console.log('SETTIMEOUT', id, ms, stack.split('\n')[2]);
  return id;
};
const originalClearTimeout = globalThis.clearTimeout;
globalThis.clearTimeout = function(id) {
  activeTimeouts.delete(id);
  return originalClearTimeout(id);
};
