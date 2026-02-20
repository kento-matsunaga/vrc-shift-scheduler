import { setupWorker } from 'msw/browser';
import { handlers, resetMockState } from './handlers';

let worker: ReturnType<typeof setupWorker> | null = null;

export async function startMSW() {
  if (worker) return;
  resetMockState();
  worker = setupWorker(...handlers);
  await worker.start({
    onUnhandledRequest: 'bypass',
    serviceWorker: {
      url: '/mockServiceWorker.js',
    },
  });
}

export function stopMSW() {
  if (worker) {
    worker.stop();
    worker = null;
  }
}
