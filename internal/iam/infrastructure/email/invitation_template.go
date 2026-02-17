package email

// invitationHTMLTemplate is the HTML template for invitation emails.
// Variables: .TenantName, .RoleName, .InviterName, .AcceptURL, .ExpiresIn
const invitationHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Invitation to {{.TenantName}}</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f5f7;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f5f7;padding:40px 0;">
<tr><td align="center">

<!-- Main Card -->
<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">

<!-- Header -->
<tr>
<td style="background:linear-gradient(135deg,#4F46E5,#7C3AED);padding:32px 40px;text-align:center;">
  <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:600;letter-spacing:-0.3px;">
    You're Invited!
  </h1>
</td>
</tr>

<!-- Body -->
<tr>
<td style="padding:36px 40px 20px;">
  <p style="margin:0 0 24px;color:#374151;font-size:15px;line-height:1.65;">
    Hi there,
  </p>
  <p style="margin:0 0 24px;color:#374151;font-size:15px;line-height:1.65;">
    <strong style="color:#111827;">{{.InviterName}}</strong> has invited you to join
    <strong style="color:#111827;">{{.TenantName}}</strong> as a
    <span style="display:inline-block;background:#EEF2FF;color:#4F46E5;padding:2px 10px;border-radius:12px;font-size:13px;font-weight:600;">{{.RoleName}}</span>.
  </p>
  <p style="margin:0 0 32px;color:#374151;font-size:15px;line-height:1.65;">
    Click the button below to accept the invitation and get started.
  </p>

  <!-- CTA Button -->
  <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto;">
  <tr>
  <td style="border-radius:8px;background:linear-gradient(135deg,#4F46E5,#7C3AED);">
    <a href="{{.AcceptURL}}"
       target="_blank"
       style="display:inline-block;padding:14px 40px;color:#ffffff;font-size:15px;font-weight:600;text-decoration:none;letter-spacing:0.3px;">
      Accept Invitation →
    </a>
  </td>
  </tr>
  </table>
</td>
</tr>

<!-- Divider -->
<tr>
<td style="padding:0 40px;">
  <hr style="border:none;border-top:1px solid #E5E7EB;margin:16px 0;">
</td>
</tr>

<!-- Footer Note -->
<tr>
<td style="padding:12px 40px 32px;">
  <p style="margin:0 0 8px;color:#9CA3AF;font-size:12px;line-height:1.5;">
    This invitation expires in <strong>{{.ExpiresIn}}</strong>.
    If you didn't expect this email, you can safely ignore it.
  </p>
  <p style="margin:0;color:#9CA3AF;font-size:12px;line-height:1.5;">
    Can't click the button? Copy this link:<br>
    <a href="{{.AcceptURL}}" style="word-break:break-all;color:#6366F1;font-size:11px;">{{.AcceptURL}}</a>
  </p>
</td>
</tr>

</table>
<!-- End Main Card -->

<!-- Brand Footer -->
<table role="presentation" width="560" cellpadding="0" cellspacing="0">
<tr>
<td style="padding:24px 40px;text-align:center;">
  <p style="margin:0;color:#9CA3AF;font-size:11px;">
    Powered by <strong style="color:#6B7280;">Converda</strong> · Omnichannel Customer Service Platform
  </p>
</td>
</tr>
</table>

</td></tr>
</table>
</body>
</html>`

// invitationTextTemplate is the plain-text fallback for invitation emails.
const invitationTextTemplate = `You're Invited!

Hi there,

{{.InviterName}} has invited you to join {{.TenantName}} as a {{.RoleName}}.

Accept the invitation by visiting this link:
{{.AcceptURL}}

This invitation expires in {{.ExpiresIn}}.
If you didn't expect this email, you can safely ignore it.

— Converda Team`
