import { useState, useEffect } from 'react';
import { changePassword, changeEmail } from '../../lib/api';
import { ApiClientError } from '../../lib/apiClient';

export function AccountSettings() {
  // Password change state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [changingPassword, setChangingPassword] = useState(false);
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // Email change state
  const [emailCurrentPassword, setEmailCurrentPassword] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [confirmNewEmail, setConfirmNewEmail] = useState('');
  const [changingEmail, setChangingEmail] = useState(false);
  const [emailError, setEmailError] = useState('');
  const [emailSuccess, setEmailSuccess] = useState('');

  // REV-002: Auto-clear success/error messages with cleanup
  useEffect(() => {
    if (passwordSuccess) {
      const timer = setTimeout(() => setPasswordSuccess(''), 3000);
      return () => clearTimeout(timer);
    }
  }, [passwordSuccess]);

  useEffect(() => {
    if (passwordError) {
      const timer = setTimeout(() => setPasswordError(''), 5000);
      return () => clearTimeout(timer);
    }
  }, [passwordError]);

  useEffect(() => {
    if (emailSuccess) {
      const timer = setTimeout(() => setEmailSuccess(''), 3000);
      return () => clearTimeout(timer);
    }
  }, [emailSuccess]);

  useEffect(() => {
    if (emailError) {
      const timer = setTimeout(() => setEmailError(''), 5000);
      return () => clearTimeout(timer);
    }
  }, [emailError]);

  // Password change handler
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    // Validation
    if (!currentPassword) {
      setPasswordError('現在のパスワードを入力してください');
      return;
    }
    if (!newPassword) {
      setPasswordError('新しいパスワードを入力してください');
      return;
    }
    if (newPassword.length < 8) {
      setPasswordError('新しいパスワードは8文字以上で入力してください');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('新しいパスワードと確認用パスワードが一致しません');
      return;
    }
    if (currentPassword === newPassword) {
      setPasswordError('新しいパスワードは現在のパスワードと異なるものを入力してください');
      return;
    }

    setChangingPassword(true);

    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
        confirm_new_password: confirmPassword,
      });
      setPasswordSuccess('パスワードを変更しました');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.message.includes('incorrect') || err.message.includes('Unauthorized')) {
          setPasswordError('現在のパスワードが正しくありません');
        } else {
          setPasswordError(err.getUserMessage());
        }
      } else {
        setPasswordError('パスワードの変更に失敗しました');
      }
      console.error('Failed to change password:', err);
    } finally {
      setChangingPassword(false);
    }
  };

  // Email change handler
  const handleChangeEmail = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    setEmailSuccess('');

    // Validation
    if (!emailCurrentPassword) {
      setEmailError('現在のパスワードを入力してください');
      return;
    }
    if (!newEmail) {
      setEmailError('新しいメールアドレスを入力してください');
      return;
    }
    if (!confirmNewEmail) {
      setEmailError('確認用メールアドレスを入力してください');
      return;
    }
    if (newEmail !== confirmNewEmail) {
      setEmailError('新しいメールアドレスと確認用メールアドレスが一致しません');
      return;
    }
    if (newEmail.length > 255) {
      setEmailError('メールアドレスは255文字以内で入力してください');
      return;
    }

    setChangingEmail(true);

    try {
      await changeEmail({
        current_password: emailCurrentPassword,
        new_email: newEmail,
        confirm_new_email: confirmNewEmail,
      });
      setEmailSuccess('メールアドレスを変更しました。次回ログイン時から新しいメールアドレスをご使用ください。');
      setEmailCurrentPassword('');
      setNewEmail('');
      setConfirmNewEmail('');
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.message.includes('incorrect') || err.message.includes('Unauthorized')) {
          setEmailError('現在のパスワードが正しくありません');
        } else if (err.message.includes('already') || err.message.includes('Conflict')) {
          setEmailError('このメールアドレスは既に使用されています');
        } else {
          setEmailError(err.getUserMessage());
        }
      } else {
        setEmailError('メールアドレスの変更に失敗しました');
      }
      console.error('Failed to change email:', err);
    } finally {
      setChangingEmail(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* パスワード変更セクション */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
          </svg>
          パスワード変更
        </h2>

        {passwordError && (
          <div role="alert" className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{passwordError}</p>
          </div>
        )}

        {passwordSuccess && (
          <div role="status" className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{passwordSuccess}</p>
          </div>
        )}

        <form onSubmit={handleChangePassword} className="space-y-4">
          <div>
            <label htmlFor="currentPassword" className="block text-sm font-medium text-gray-700 mb-1">
              現在のパスワード
            </label>
            <input
              type="password"
              id="currentPassword"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="current-password"
            />
          </div>

          <div>
            <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
              新しいパスワード
            </label>
            <input
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="new-password"
            />
            <p className="text-xs text-gray-500 mt-1">8文字以上で入力してください</p>
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              新しいパスワード（確認）
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="input-field"
              disabled={changingPassword}
              autoComplete="new-password"
            />
          </div>

          <button
            type="submit"
            disabled={changingPassword || !currentPassword || !newPassword || !confirmPassword}
            className="btn-primary"
          >
            {changingPassword ? 'パスワード変更中...' : 'パスワードを変更'}
          </button>
        </form>
      </div>

      {/* メールアドレス変更セクション */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          メールアドレス変更
        </h2>

        {emailError && (
          <div role="alert" className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{emailError}</p>
          </div>
        )}

        {emailSuccess && (
          <div role="status" className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-green-800">{emailSuccess}</p>
          </div>
        )}

        <form onSubmit={handleChangeEmail} className="space-y-4">
          <div>
            <label htmlFor="emailCurrentPassword" className="block text-sm font-medium text-gray-700 mb-1">
              現在のパスワード
            </label>
            <input
              type="password"
              id="emailCurrentPassword"
              value={emailCurrentPassword}
              onChange={(e) => setEmailCurrentPassword(e.target.value)}
              className="input-field"
              disabled={changingEmail}
              autoComplete="current-password"
            />
          </div>

          <div>
            <label htmlFor="newEmail" className="block text-sm font-medium text-gray-700 mb-1">
              新しいメールアドレス
            </label>
            <input
              type="email"
              id="newEmail"
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              className="input-field"
              disabled={changingEmail}
              autoComplete="email"
            />
          </div>

          <div>
            <label htmlFor="confirmNewEmail" className="block text-sm font-medium text-gray-700 mb-1">
              新しいメールアドレス（確認）
            </label>
            <input
              type="email"
              id="confirmNewEmail"
              value={confirmNewEmail}
              onChange={(e) => setConfirmNewEmail(e.target.value)}
              className="input-field"
              disabled={changingEmail}
              autoComplete="email"
            />
          </div>

          <button
            type="submit"
            disabled={changingEmail || !emailCurrentPassword || !newEmail || !confirmNewEmail}
            className="btn-primary"
          >
            {changingEmail ? 'メールアドレス変更中...' : 'メールアドレスを変更'}
          </button>
        </form>
      </div>
    </div>
  );
}
