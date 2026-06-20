export function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, reject, resolve };
}

export function snapshot(store) {
  let value;
  const unsubscribe = store.subscribe((next) => {
    value = next;
  });
  unsubscribe();
  return value;
}
