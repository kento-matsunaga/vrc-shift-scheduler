import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useOnboardingContext } from '../OnboardingContext';
import type { OnboardingPhase } from '../steps/types';

/**
 * React controlled input への値設定
 * React 内部の state と DOM の value を同期させるため、
 * native setter を呼び出した後に input イベントを発行する
 */
export function setReactInputValue(el: HTMLInputElement | HTMLTextAreaElement, value: string) {
  const proto = el instanceof HTMLTextAreaElement
    ? HTMLTextAreaElement.prototype
    : HTMLInputElement.prototype;
  const setter = Object.getOwnPropertyDescriptor(proto, 'value')?.set;
  setter?.call(el, value);
  el.dispatchEvent(new Event('input', { bubbles: true }));
  el.dispatchEvent(new Event('change', { bubbles: true }));
}

/**
 * select 要素への値設定
 */
export function setReactSelectValue(el: HTMLSelectElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value')?.set;
  setter?.call(el, value);
  el.dispatchEvent(new Event('change', { bubbles: true }));
}

/**
 * DOM要素の出現を待つ（MutationObserver）
 */
export function waitForElement(selector: string, timeout = 3000): Promise<Element | null> {
  return new Promise((resolve) => {
    const existing = document.querySelector(selector);
    if (existing) {
      resolve(existing);
      return;
    }

    const observer = new MutationObserver(() => {
      const el = document.querySelector(selector);
      if (el) {
        observer.disconnect();
        resolve(el);
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true,
    });

    setTimeout(() => {
      observer.disconnect();
      resolve(null);
    }, timeout);
  });
}

/**
 * 指定時間待機
 */
export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * メインのオンボーディングフック
 */
export function useOnboarding() {
  const { state, startTour, stopTour, setPhase, nextPhase } = useOnboardingContext();
  const navigate = useNavigate();

  const navigateAndWait = useCallback(async (path: string, waitMs = 300) => {
    navigate(path);
    await delay(waitMs);
  }, [navigate]);

  const goToPhase = useCallback(async (phase: OnboardingPhase, path?: string) => {
    if (path) {
      await navigateAndWait(path);
    }
    setPhase(phase);
  }, [setPhase, navigateAndWait]);

  return {
    state,
    startTour,
    stopTour,
    setPhase,
    nextPhase,
    navigate: navigateAndWait,
    goToPhase,
  };
}
