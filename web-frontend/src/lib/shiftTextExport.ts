/**
 * シフト配置をテキスト形式でエクスポートするユーティリティ
 * Issue #85: 名前のみアウトプット機能（インスタンス表用）
 */

export type MemberSeparator = 'newline' | 'comma';

export interface SlotAssignment {
  memberName: string;
}

export interface SlotData {
  slotName: string;
  assignments: SlotAssignment[];
}

export interface InstanceData {
  instanceName: string;
  slots: SlotData[];
}

/**
 * シフト配置データをテキスト形式に変換
 * @param instances インスタンスごとのシフトデータ
 * @param separator メンバー名の区切り文字（'newline' or 'comma'）
 * @returns テキスト形式の文字列
 */
export function generateShiftText(
  instances: InstanceData[],
  separator: MemberSeparator = 'newline'
): string {
  const lines: string[] = [];

  instances.forEach((instance, index) => {
    // インスタンス間に空行を入れる（最初以外）
    if (index > 0) {
      lines.push('');
    }

    // インスタンス名
    lines.push(instance.instanceName);

    // 各役職とメンバー
    instance.slots.forEach((slot) => {
      // 役職名
      lines.push(slot.slotName);

      // メンバー名
      const memberNames = slot.assignments.map((a) => a.memberName);
      if (memberNames.length > 0) {
        if (separator === 'comma') {
          lines.push(memberNames.join(', '));
        } else {
          memberNames.forEach((name) => lines.push(name));
        }
      }
    });
  });

  return lines.join('\n');
}

/**
 * テキストをクリップボードにコピー
 * Clipboard APIが利用できない環境（HTTP、古いブラウザ）ではfallbackを使用
 * @param text コピーするテキスト
 * @returns 成功したかどうか
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    // Clipboard APIが利用可能な場合
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }

    // Fallback: execCommandを使用（HTTP環境や古いブラウザ向け）
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-9999px';
    textArea.style.top = '-9999px';
    textArea.style.opacity = '0';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    return successful;
  } catch (err) {
    console.error('Failed to copy to clipboard:', err);
    return false;
  }
}
