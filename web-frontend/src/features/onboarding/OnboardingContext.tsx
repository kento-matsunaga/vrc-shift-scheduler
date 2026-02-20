import { createContext, useContext, useReducer, useCallback, useEffect, type ReactNode } from 'react';
import { DUMMY_IDS, PHASE_ORDER, type OnboardingState, type OnboardingAction, type OnboardingPhase } from './steps/types';

const STORAGE_KEY = 'vrcshift_onboarding';

const initialState: OnboardingState = {
  isActive: false,
  currentPhase: 'idle',
  dummyIds: DUMMY_IDS,
  mswReady: false,
};

function reducer(state: OnboardingState, action: OnboardingAction): OnboardingState {
  switch (action.type) {
    case 'START':
      return { ...state, isActive: true, currentPhase: 'sidebar' };
    case 'STOP':
      return { ...initialState };
    case 'SET_PHASE':
      return { ...state, currentPhase: action.phase };
    case 'NEXT_PHASE': {
      const currentIndex = PHASE_ORDER.indexOf(state.currentPhase);
      if (currentIndex < 0 || currentIndex >= PHASE_ORDER.length - 1) {
        return { ...state, currentPhase: 'complete' };
      }
      return { ...state, currentPhase: PHASE_ORDER[currentIndex + 1] };
    }
    case 'SET_MSW_READY':
      return { ...state, mswReady: action.ready };
    case 'RESTORE':
      return action.state;
    default:
      return state;
  }
}

interface OnboardingContextValue {
  state: OnboardingState;
  startTour: () => Promise<void>;
  stopTour: () => Promise<void>;
  setPhase: (phase: OnboardingPhase) => void;
  nextPhase: () => void;
}

const OnboardingContext = createContext<OnboardingContextValue | null>(null);

export function OnboardingProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState, () => {
    // sessionStorage から復元を試みる
    try {
      const saved = sessionStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved) as OnboardingState;
        if (parsed.isActive) {
          return { ...parsed, mswReady: false };
        }
      }
    } catch {
      // パースエラーは無視
    }
    return initialState;
  });

  // sessionStorage に永続化
  useEffect(() => {
    if (state.isActive) {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } else {
      sessionStorage.removeItem(STORAGE_KEY);
    }
  }, [state]);

  const startTour = useCallback(async () => {
    try {
      const { startMSW } = await import('./mocks/browser');
      await startMSW();
      dispatch({ type: 'SET_MSW_READY', ready: true });
      dispatch({ type: 'START' });
    } catch {
      // MSW起動失敗（ブラウザ非対応等）→ ツアーは開始しない
    }
  }, []);

  const stopTour = useCallback(async () => {
    try {
      const { stopMSW } = await import('./mocks/browser');
      stopMSW();
    } catch {
      // ignore
    }
    dispatch({ type: 'STOP' });
  }, []);

  const setPhase = useCallback((phase: OnboardingPhase) => {
    dispatch({ type: 'SET_PHASE', phase });
  }, []);

  const nextPhase = useCallback(() => {
    dispatch({ type: 'NEXT_PHASE' });
  }, []);

  // MSW復元（セッションリストア時）
  useEffect(() => {
    if (!state.isActive || state.mswReady) return;
    let cancelled = false;
    (async () => {
      try {
        const { startMSW } = await import('./mocks/browser');
        await startMSW();
        if (!cancelled) {
          dispatch({ type: 'SET_MSW_READY', ready: true });
        }
      } catch {
        // MSW復元失敗時は無視（次回リロードで再試行）
      }
    })();
    return () => { cancelled = true; };
  }, [state.isActive, state.mswReady]);

  return (
    <OnboardingContext.Provider value={{ state, startTour, stopTour, setPhase, nextPhase }}>
      {children}
    </OnboardingContext.Provider>
  );
}

export function useOnboardingContext() {
  const ctx = useContext(OnboardingContext);
  if (!ctx) throw new Error('useOnboardingContext must be used within OnboardingProvider');
  return ctx;
}
