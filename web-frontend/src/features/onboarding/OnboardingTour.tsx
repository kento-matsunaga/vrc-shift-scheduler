import { useEffect, useRef, useCallback } from 'react';
import { useLocation } from 'react-router-dom';
import { driver, type DriveStep, type Config } from 'driver.js';
import 'driver.js/dist/driver.css';
import './onboarding.css';
import { useOnboarding, waitForElement, delay, setReactInputValue } from './hooks/useOnboarding';
import { DUMMY_IDS } from './steps/types';

type DriverInstance = ReturnType<typeof driver>;

export function OnboardingTour() {
  const { state, stopTour, setPhase, navigate } = useOnboarding();
  const driverRef = useRef<DriverInstance | null>(null);
  const location = useLocation();

  // ドライバー破棄
  const destroyDriver = useCallback(() => {
    if (driverRef.current) {
      driverRef.current.destroy();
      driverRef.current = null;
    }
  }, []);

  // ESCキーで中断確認
  useEffect(() => {
    if (!state.isActive) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        destroyDriver();
        // ESC 中断 → 確認はStartTutorialButtonのUIに委ねる
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [state.isActive, destroyDriver]);

  // フェーズ変更時にドライバーを再作成
  useEffect(() => {
    if (!state.isActive || !state.mswReady || state.currentPhase === 'idle') return;

    const runPhase = async () => {
      destroyDriver();
      await delay(200);

      switch (state.currentPhase) {
        case 'sidebar':
          await runSidebarPhase();
          break;
        case 'role':
          await runRolePhase();
          break;
        case 'event':
          await runEventPhase();
          break;
        case 'template':
          await runTemplatePhase();
          break;
        case 'businessDay':
          await runBusinessDayPhase();
          break;
        case 'shiftSlot':
          await runShiftSlotPhase();
          break;
        case 'member':
          await runMemberPhase();
          break;
        case 'attendance':
          await runAttendancePhase();
          break;
        case 'attendanceResponse':
          await runAttendanceResponsePhase();
          break;
        case 'attendanceDetail':
          await runAttendanceDetailPhase();
          break;
        case 'shiftAdjustment':
          await runShiftAdjustmentPhase();
          break;
        case 'calendar':
          await runCalendarPhase();
          break;
        case 'summary':
          await runSummaryPhase();
          break;
        case 'complete':
          await runCompletePhase();
          break;
      }
    };

    runPhase();

    return () => destroyDriver();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- フェーズ変更のみトリガー
  }, [state.currentPhase, state.isActive, state.mswReady]);

  // ドライバー作成ヘルパー
  function createDriver(steps: DriveStep[], onComplete: () => void, config?: Partial<Config>) {
    const d = driver({
      showProgress: true,
      animate: true,
      allowClose: false,
      overlayClickBehavior: () => { /* overlay クリック無効化 */ },
      stagePadding: 8,
      stageRadius: 8,
      steps,
      onDestroyStarted: () => {
        d.destroy();
        // 最後のステップなら完了処理（次フェーズへ遷移）
        if (d.isLastStep()) {
          onComplete();
        }
      },
      ...config,
    });
    driverRef.current = d;
    d.drive();
    return d;
  }

  // 要素クリックをシミュレート
  function clickElement(selector: string) {
    const el = document.querySelector(selector);
    if (el instanceof HTMLElement) {
      el.click();
    }
  }

  // === Phase implementations ===

  async function runSidebarPhase() {
    if (!location.pathname.startsWith('/events')) {
      await navigate('/events');
    }

    createDriver([
      {
        popover: {
          title: 'VRC Shift Scheduler へようこそ！',
          description: 'このツアーでは、シフト管理の全機能を体験します。ダミーデータを使うので実データには影響しません。',
        },
      },
      {
        element: '#nav-events',
        popover: {
          title: 'イベント',
          description: 'イベント（バーやクラブなど）を管理します。イベントごとに営業日やシフトを設定できます。',
        },
      },
      {
        element: '#nav-members',
        popover: {
          title: 'メンバー',
          description: 'シフトに参加するメンバーを管理します。ロールやグループで分類できます。',
        },
      },
      {
        element: '#nav-roles',
        popover: {
          title: 'ロール',
          description: 'バーテンダーやMCなど、メンバーの役割を定義します。次のステップで実際に作成してみましょう。',
        },
      },
      {
        element: '#nav-attendance',
        popover: {
          title: '出欠確認',
          description: 'メンバーに出欠を確認し、回答を集計します。URLを送るだけで簡単に回答してもらえます。',
        },
      },
      {
        element: '#nav-calendars',
        popover: {
          title: 'カレンダー',
          description: 'シフト予定を外部共有できるカレンダーです。では、まずロールを作成しましょう！',
        },
      },
    ], () => {
      setPhase('role');
    });
  }

  async function runRolePhase() {
    await navigate('/roles');
    await delay(500);

    createDriver([
      {
        element: '#btn-create-role',
        popover: {
          title: 'ロールを作成しましょう',
          description: '「バーテンダー」ロールを作成します。ボタンをクリックしてください。',
        },
        onHighlightStarted: () => {
          // 自動で次のステップでクリックを実行
        },
      },
      {
        popover: {
          title: 'ロール作成',
          description: 'ロール作成モーダルを開きます...',
          onNextClick: async () => {
            clickElement('#btn-create-role');
            await waitForElement('.fixed.inset-0'); // モーダル待ち
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#name',
        popover: {
          title: 'ロール名を入力',
          description: '「バーテンダー」と入力します。',
          onNextClick: async () => {
            const input = document.querySelector('#name') as HTMLInputElement;
            if (input) setReactInputValue(input, 'バーテンダー');
            await delay(200);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#btn-submit-role',
        popover: {
          title: '作成！',
          description: '作成ボタンをクリックして、ロールを登録します。',
          onNextClick: async () => {
            clickElement('#btn-submit-role');
            await delay(800);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'ロール作成完了！',
          description: 'バーテンダーロールが作成されました。次はイベントを作成しましょう！',
        },
      },
    ], () => {
      setPhase('event');
    });
  }

  async function runEventPhase() {
    await navigate('/events');
    await delay(500);

    createDriver([
      {
        element: '#btn-create-event',
        popover: {
          title: 'イベントを作成',
          description: '「チュートリアル Bar」イベントを作成します。',
          onNextClick: async () => {
            clickElement('#btn-create-event');
            await waitForElement('.fixed.inset-0');
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#eventName',
        popover: {
          title: 'イベント名',
          description: '「チュートリアル Bar」と入力します。',
          onNextClick: async () => {
            const input = document.querySelector('#eventName') as HTMLInputElement;
            if (input) setReactInputValue(input, 'チュートリアル Bar');
            await delay(200);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#btn-submit-event',
        popover: {
          title: '作成！',
          description: 'イベントを作成します。',
          onNextClick: async () => {
            clickElement('#btn-submit-event');
            await delay(800);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'イベント作成完了！',
          description: 'チュートリアル Bar が作成されました。次はシフトテンプレートを作成しましょう。',
        },
      },
    ], () => {
      setPhase('template');
    });
  }

  async function runTemplatePhase() {
    await navigate(`/events/${DUMMY_IDS.eventId}/business-days`);
    await delay(500);

    createDriver([
      {
        element: '#link-template-management',
        popover: {
          title: 'テンプレート管理',
          description: 'テンプレートを使うと、シフト枠の構成を再利用できます。テンプレート管理ページへ移動しましょう。',
          onNextClick: async () => {
            await navigate(`/events/${DUMMY_IDS.eventId}/templates`);
            await delay(500);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#btn-create-template',
        popover: {
          title: '新規テンプレート作成',
          description: 'テンプレートを作成します。',
          onNextClick: async () => {
            await navigate(`/events/${DUMMY_IDS.eventId}/templates/new`);
            await delay(500);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#templateName',
        popover: {
          title: 'テンプレート名',
          description: '「メインインスタンス構成」と入力します。',
          onNextClick: async () => {
            const input = document.querySelector('#templateName') as HTMLInputElement;
            if (input) setReactInputValue(input, 'メインインスタンス構成');
            await delay(200);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'テンプレートの仕組み',
          description: 'テンプレートには「インスタンス」と「役職」を定義します。インスタンスはVRChatのワールドインスタンスに対応し、役職はバーテンダーやMCなどのシフト枠です。\n\nここでは、MSWがテンプレート作成をシミュレートします。',
          onNextClick: async () => {
            // MSWがテンプレート作成をインターセプト
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'テンプレート作成完了！',
          description: '「メインインスタンス構成」テンプレートが作成されました。メインフロアにバーテンダー2名、MC1名の構成です。\n\n次は営業日を作成し、このテンプレートを適用しましょう。',
        },
      },
    ], () => {
      setPhase('businessDay');
    });
  }

  async function runBusinessDayPhase() {
    await navigate(`/events/${DUMMY_IDS.eventId}/business-days`);
    await delay(500);

    createDriver([
      {
        element: '#btn-create-business-day',
        popover: {
          title: '営業日を追加',
          description: '営業日を作成します。テンプレートを選択すると、シフト枠が自動的に作成されます。',
          onNextClick: async () => {
            clickElement('#btn-create-business-day');
            await waitForElement('.fixed.inset-0');
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'テンプレートを選択',
          description: '先ほど作成した「メインインスタンス構成」テンプレートを選択します。テンプレートを選ぶと、バーテンダー2名・MC1名のシフト枠が自動作成されます。',
          onNextClick: async () => {
            // MSWが営業日+シフト枠をインターセプト
            await delay(500);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: '営業日作成完了！',
          description: '営業日が作成され、テンプレートからシフト枠が自動作成されました。シフト枠を確認しましょう。',
        },
      },
    ], () => {
      setPhase('shiftSlot');
    });
  }

  async function runShiftSlotPhase() {
    await navigate(`/business-days/${DUMMY_IDS.businessDayId}/shift-slots`);
    await delay(500);

    createDriver([
      {
        popover: {
          title: 'シフト枠一覧',
          description: 'テンプレートから自動作成されたシフト枠です。「メインフロア」インスタンスに、バーテンダー（2名）とMC（1名）の枠があります。',
        },
      },
      {
        popover: {
          title: 'シフト枠の構造',
          description: 'シフト枠はインスタンスごとにグループ化されています。各枠には役職名、必要人数、時間帯が設定されています。\n\nテンプレートを使えば、毎回同じ構成を簡単に適用できます。',
        },
      },
      {
        popover: {
          title: '次はメンバー追加',
          description: 'シフト枠が準備できました。次はメンバーを追加して、シフトに割り当てましょう。',
        },
      },
    ], () => {
      setPhase('member');
    });
  }

  async function runMemberPhase() {
    await navigate('/members');
    await delay(500);

    createDriver([
      {
        element: '#btn-create-member',
        popover: {
          title: 'メンバーを追加',
          description: '3名のメンバーを追加します。',
          onNextClick: async () => {
            clickElement('#btn-create-member');
            await waitForElement('.fixed.inset-0');
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#member-display-name',
        popover: {
          title: 'メンバー名を入力',
          description: '「田中太郎」と入力します。',
          onNextClick: async () => {
            const input = document.querySelector('#member-display-name') as HTMLInputElement;
            if (input) setReactInputValue(input, '田中太郎');
            await delay(200);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        element: '#btn-submit-member',
        popover: {
          title: '登録！',
          description: 'メンバーを登録します。',
          onNextClick: async () => {
            clickElement('#btn-submit-member');
            await delay(800);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'メンバー追加完了！',
          description: '3名のメンバー（田中太郎、佐藤花子、鈴木一郎）が追加されました。各メンバーにロールも割り当て済みです。\n\n次は出欠確認を作成しましょう！',
        },
      },
    ], () => {
      setPhase('attendance');
    });
  }

  async function runAttendancePhase() {
    await navigate('/attendance');
    await delay(500);

    createDriver([
      {
        element: '#btn-create-attendance',
        popover: {
          title: '出欠確認を作成',
          description: 'メンバーに出欠を確認するフォームを作成します。',
          onNextClick: async () => {
            clickElement('#btn-create-attendance');
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: '出欠確認の仕組み',
          description: '出欠確認を作成すると、公開URLが発行されます。このURLをメンバーに送信すると、各日の参加/不参加を回答してもらえます。\n\nイベントから日程を取り込むこともでき、締切も設定できます。',
          onNextClick: async () => {
            // MSWが出欠確認作成をインターセプト
            await delay(500);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: '出欠確認作成完了！',
          description: '出欠確認が作成され、公開URLが発行されました。\n\n次のステップでは、メンバーがどのように回答するかを説明します。',
        },
      },
    ], () => {
      setPhase('attendanceResponse');
    });
  }

  async function runAttendanceResponsePhase() {
    createDriver([
      {
        popover: {
          title: '出欠の回答方法',
          description: 'メンバーは公開URLを開いて、自分の名前を選択し、各日程の参加/不参加を回答します。\n\n参加可能時間帯やメモも入力できます。',
        },
      },
      {
        popover: {
          title: '回答が集まりました',
          description: '3名全員が「参加」と回答しました。回答状況を詳細画面で確認しましょう。',
        },
      },
    ], () => {
      setPhase('attendanceDetail');
    });
  }

  async function runAttendanceDetailPhase() {
    await navigate(`/attendance/${DUMMY_IDS.collectionId}`);
    await delay(500);

    createDriver([
      {
        popover: {
          title: '出欠確認詳細',
          description: '回答状況を一覧で確認できます。各メンバーの参加/不参加が表示されています。',
        },
      },
      {
        element: '#btn-close-collection',
        popover: {
          title: '締め切り',
          description: '「締め切る」ボタンで出欠の受付を終了できます。締切後は回答の変更ができなくなります。',
        },
      },
      {
        element: '#btn-shift-adjustment',
        popover: {
          title: 'シフト調整へ',
          description: '出欠結果をもとに、シフトの割り当てを行います。次のステップで体験しましょう！',
        },
      },
    ], () => {
      setPhase('shiftAdjustment');
    });
  }

  async function runShiftAdjustmentPhase() {
    await navigate(`/attendance/${DUMMY_IDS.collectionId}/shift-adjustment`);
    await delay(500);

    createDriver([
      {
        popover: {
          title: 'シフト調整',
          description: '出欠確認の結果をもとに、各シフト枠にメンバーを割り当てます。',
        },
      },
      {
        element: '#shift-date-tabs',
        popover: {
          title: '日付タブ',
          description: '複数日程がある場合、タブで切り替えられます。',
        },
      },
      {
        element: '#shift-attending-members',
        popover: {
          title: '参加メンバー',
          description: '「参加」と回答したメンバーの一覧です。ここからシフト枠に割り当てます。',
        },
      },
      {
        element: '#shift-slot-assignments',
        popover: {
          title: 'シフト枠割り当て',
          description: 'テンプレートで定義した枠（バーテンダー2名、MC1名）にメンバーを割り当てます。\n\nドロップダウンからメンバーを選んで「追加」ボタンで割り当てます。',
        },
      },
      {
        popover: {
          title: 'シフト調整完了！',
          description: 'シフトの割り当てができました。最後にカレンダー機能を確認しましょう！',
        },
      },
    ], () => {
      setPhase('calendar');
    });
  }

  async function runCalendarPhase() {
    await navigate('/calendars');
    await delay(500);

    createDriver([
      {
        element: '#btn-create-calendar',
        popover: {
          title: 'カレンダーを作成',
          description: 'カレンダーを作成して、シフト予定を外部共有できます。',
          onNextClick: async () => {
            clickElement('#btn-create-calendar');
            await waitForElement('.fixed.inset-0');
            await delay(300);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'カレンダーの仕組み',
          description: 'カレンダーにはイベントを紐付けます。公開URLを共有すると、メンバーがシフト予定を確認できます。',
          onNextClick: async () => {
            // MSWがカレンダー作成をインターセプト
            await delay(500);
            driverRef.current?.moveNext();
          },
        },
      },
      {
        popover: {
          title: 'カレンダー作成完了！',
          description: '公開URLを共有すると、誰でもシフト予定を閲覧できます。\n\nこれで全機能の体験が完了しました！',
        },
      },
    ], () => {
      setPhase('summary');
    });
  }

  async function runSummaryPhase() {
    createDriver([
      {
        popover: {
          title: '全機能まとめ',
          description: [
            '体験した機能の振り返り:',
            '',
            '1. ロール: メンバーの役割を定義（バーテンダー、MC等）',
            '2. テンプレート: シフト構成を再利用可能な形で保存',
            '3. イベント → 営業日 → シフト枠: 階層構造でシフトを管理',
            '4. 出欠確認 → シフト調整: 出欠を集計してシフトに反映',
            '5. カレンダー: シフト予定を外部共有',
          ].join('\n'),
        },
      },
    ], () => {
      setPhase('complete');
    });
  }

  async function runCompletePhase() {
    createDriver([
      {
        popover: {
          title: 'チュートリアル完了！',
          description: 'お疲れ様でした！全ての機能を体験できました。\n\nこの画面を閉じるとダミーデータは自動的にクリーンアップされます。\n\nもう一度体験したい場合は、ヘッダーの「体験ツアー」ボタンから再開できます。',
        },
      },
    ], async () => {
      await stopTour();
      window.location.reload();
    });
  }

  // チュートリアルがアクティブでないときは何も表示しない
  if (!state.isActive) return null;

  return null; // driver.jsはDOM操作で表示するため、React要素は不要
}
