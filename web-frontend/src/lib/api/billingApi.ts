// 公開API（ライセンスクレーム用）
// 管理者APIは admin-frontend に移動しました

export async function claimLicense(request: {
  email: string;
  password: string;
  display_name: string;
  tenant_name: string;
  license_key: string;
}): Promise<{
  data: {
    tenant_id: string;
    admin_id: string;
    tenant_name: string;
    display_name: string;
    email: string;
    message: string;
  };
}> {
  const response = await fetch('/api/v1/public/license/claim', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error?.message || 'ライセンスの登録に失敗しました');
  }

  return response.json();
}
