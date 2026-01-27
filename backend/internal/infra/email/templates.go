package email

import (
	"bytes"
	"html/template"
	"sync"
	"time"
)

// Cached templates (parsed once at first use)
var (
	htmlTemplateOnce sync.Once
	textTemplateOnce sync.Once
	cachedHTMLTmpl   *template.Template
	cachedTextTmpl   *template.Template
)

// InvitationEmailData represents data for the invitation email template
type InvitationEmailData struct {
	InviterName   string
	TenantName    string
	Role          string
	RoleJapanese  string
	ExpiresAt     string
	InvitationURL string
}

// invitationHTMLTemplate is the HTML template for invitation emails
const invitationHTMLTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>管理者招待</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Helvetica Neue', Arial, 'Hiragino Kaku Gothic ProN', 'Hiragino Sans', Meiryo, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 100%; border-collapse: collapse; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 32px 40px; background-color: #4F46E5; border-radius: 8px 8px 0 0;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 24px; font-weight: 600;">VRC Shift Scheduler</h1>
                        </td>
                    </tr>
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px;">
                            <p style="margin: 0 0 24px; font-size: 16px; line-height: 1.6; color: #333333;">
                                こんにちは、
                            </p>
                            <p style="margin: 0 0 24px; font-size: 16px; line-height: 1.6; color: #333333;">
                                <strong>{{.InviterName}}</strong> さんから「<strong>{{.TenantName}}</strong>」の管理者として招待されました。
                            </p>

                            <!-- Invitation Details -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 24px 0; background-color: #f8f9fa; border-radius: 6px;">
                                <tr>
                                    <td style="padding: 20px;">
                                        <table role="presentation" style="width: 100%; border-collapse: collapse;">
                                            <tr>
                                                <td style="padding: 8px 0; color: #666666; font-size: 14px;">ロール</td>
                                                <td style="padding: 8px 0; color: #333333; font-size: 14px; font-weight: 600;">{{.RoleJapanese}}</td>
                                            </tr>
                                            <tr>
                                                <td style="padding: 8px 0; color: #666666; font-size: 14px;">有効期限</td>
                                                <td style="padding: 8px 0; color: #333333; font-size: 14px; font-weight: 600;">{{.ExpiresAt}}</td>
                                            </tr>
                                        </table>
                                    </td>
                                </tr>
                            </table>

                            <!-- CTA Button -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 32px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="{{.InvitationURL}}" style="display: inline-block; padding: 16px 48px; background-color: #4F46E5; color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; border-radius: 6px;">
                                            招待を受諾する
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 24px 0 0; font-size: 14px; line-height: 1.6; color: #666666;">
                                ※ このリンクは {{.ExpiresAt}} まで有効です。<br>
                                ※ 心当たりがない場合は、このメールを無視してください。
                            </p>
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td style="padding: 24px 40px; background-color: #f8f9fa; border-radius: 0 0 8px 8px; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0; font-size: 12px; color: #9ca3af; text-align: center;">
                                このメールは VRC Shift Scheduler から自動送信されています。<br>
                                <a href="https://vrcshift.com" style="color: #4F46E5; text-decoration: none;">https://vrcshift.com</a>
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`

// invitationTextTemplate is the plain text template for invitation emails
const invitationTextTemplate = `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VRC Shift Scheduler
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

こんにちは、

{{.InviterName}} さんから「{{.TenantName}}」の管理者として
招待されました。

■ 招待内容
  ロール: {{.RoleJapanese}}
  有効期限: {{.ExpiresAt}}

下記のリンクから招待を受諾し、アカウントを作成してください。

{{.InvitationURL}}

※ このリンクは {{.ExpiresAt}} まで有効です。
※ 心当たりがない場合は、このメールを無視してください。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
このメールは VRC Shift Scheduler から自動送信されています。
https://vrcshift.com
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

// RoleToJapanese converts a role string to Japanese
func RoleToJapanese(role string) string {
	switch role {
	case "owner":
		return "オーナー"
	case "manager":
		return "マネージャー"
	default:
		return role
	}
}

// RenderInvitationHTML renders the HTML version of the invitation email
func RenderInvitationHTML(data InvitationEmailData) (string, error) {
	htmlTemplateOnce.Do(func() {
		cachedHTMLTmpl = template.Must(template.New("invitation_html").Parse(invitationHTMLTemplate))
	})

	var buf bytes.Buffer
	if err := cachedHTMLTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderInvitationText renders the plain text version of the invitation email
func RenderInvitationText(data InvitationEmailData) (string, error) {
	textTemplateOnce.Do(func() {
		cachedTextTmpl = template.Must(template.New("invitation_text").Parse(invitationTextTemplate))
	})

	var buf bytes.Buffer
	if err := cachedTextTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// FormatExpiresAt formats the expiration time for display
func FormatExpiresAt(t time.Time) string {
	return t.In(time.FixedZone("JST", 9*60*60)).Format("2006年1月2日 15:04")
}

// PasswordResetEmailData represents data for the password reset email template
type PasswordResetEmailData struct {
	ResetURL  string
	ExpiresAt string
}

// Cached password reset templates
var (
	passwordResetHTMLTemplateOnce sync.Once
	passwordResetTextTemplateOnce sync.Once
	cachedPasswordResetHTMLTmpl   *template.Template
	cachedPasswordResetTextTmpl   *template.Template
)

// passwordResetHTMLTemplate is the HTML template for password reset emails
const passwordResetHTMLTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>パスワードリセット</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Helvetica Neue', Arial, 'Hiragino Kaku Gothic ProN', 'Hiragino Sans', Meiryo, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 100%; border-collapse: collapse; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 32px 40px; background-color: #4F46E5; border-radius: 8px 8px 0 0;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 24px; font-weight: 600;">VRC Shift Scheduler</h1>
                        </td>
                    </tr>
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 24px; font-size: 20px; color: #333333;">パスワードリセットのリクエスト</h2>
                            <p style="margin: 0 0 24px; font-size: 16px; line-height: 1.6; color: #333333;">
                                パスワードリセットのリクエストを受け付けました。<br>
                                下記のボタンをクリックして、新しいパスワードを設定してください。
                            </p>

                            <!-- CTA Button -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 32px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="{{.ResetURL}}" style="display: inline-block; padding: 16px 48px; background-color: #4F46E5; color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; border-radius: 6px;">
                                            パスワードをリセット
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <!-- Expiration Notice -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 24px 0; background-color: #fef3c7; border-radius: 6px; border-left: 4px solid #f59e0b;">
                                <tr>
                                    <td style="padding: 16px 20px;">
                                        <p style="margin: 0; font-size: 14px; color: #92400e;">
                                            <strong>有効期限:</strong> {{.ExpiresAt}}<br>
                                            このリンクは1時間のみ有効です。
                                        </p>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 24px 0 0; font-size: 14px; line-height: 1.6; color: #666666;">
                                ※ このリクエストに心当たりがない場合は、このメールを無視してください。<br>
                                ※ パスワードは変更されません。
                            </p>
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td style="padding: 24px 40px; background-color: #f8f9fa; border-radius: 0 0 8px 8px; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0; font-size: 12px; color: #9ca3af; text-align: center;">
                                このメールは VRC Shift Scheduler から自動送信されています。<br>
                                <a href="https://vrcshift.com" style="color: #4F46E5; text-decoration: none;">https://vrcshift.com</a>
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`

// passwordResetTextTemplate is the plain text template for password reset emails
const passwordResetTextTemplate = `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VRC Shift Scheduler
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

パスワードリセットのリクエスト

パスワードリセットのリクエストを受け付けました。
下記のリンクから新しいパスワードを設定してください。

{{.ResetURL}}

■ 有効期限
  {{.ExpiresAt}}
  ※ このリンクは1時間のみ有効です。

※ このリクエストに心当たりがない場合は、このメールを無視してください。
※ パスワードは変更されません。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
このメールは VRC Shift Scheduler から自動送信されています。
https://vrcshift.com
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

// RenderPasswordResetHTML renders the HTML version of the password reset email
func RenderPasswordResetHTML(data PasswordResetEmailData) (string, error) {
	passwordResetHTMLTemplateOnce.Do(func() {
		cachedPasswordResetHTMLTmpl = template.Must(template.New("password_reset_html").Parse(passwordResetHTMLTemplate))
	})

	var buf bytes.Buffer
	if err := cachedPasswordResetHTMLTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderPasswordResetText renders the plain text version of the password reset email
func RenderPasswordResetText(data PasswordResetEmailData) (string, error) {
	passwordResetTextTemplateOnce.Do(func() {
		cachedPasswordResetTextTmpl = template.Must(template.New("password_reset_text").Parse(passwordResetTextTemplate))
	})

	var buf bytes.Buffer
	if err := cachedPasswordResetTextTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
