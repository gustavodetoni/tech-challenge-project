package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
)

const defaultBrevoAPIURL = "https://api.brevo.com/v3/smtp/email"

type BrevoConfig struct {
	APIKey      string
	SenderEmail string
	SenderName  string
	APIURL      string
}

type BrevoNotifier struct {
	cfg    BrevoConfig
	client *http.Client
}

func NewBrevoNotifier(cfg BrevoConfig, client *http.Client) (*BrevoNotifier, error) {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.SenderEmail = strings.TrimSpace(cfg.SenderEmail)
	cfg.SenderName = strings.TrimSpace(cfg.SenderName)
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	if cfg.APIURL == "" {
		cfg.APIURL = defaultBrevoAPIURL
	}
	if cfg.SenderName == "" {
		cfg.SenderName = "Oficina - Pos"
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("brevo api key is required")
	}
	if cfg.SenderEmail == "" {
		return nil, fmt.Errorf("brevo sender email is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &BrevoNotifier{cfg: cfg, client: client}, nil
}

func (n *BrevoNotifier) NotifyStatusChanged(ctx context.Context, input repository.StatusChangedNotification) error {
	toEmail := strings.TrimSpace(input.ClientEmail)
	if toEmail == "" {
		return nil
	}
	toName := strings.TrimSpace(input.ClientName)
	if toName == "" {
		toName = "Cliente"
	}

	payload := brevoEmailRequest{
		Sender: brevoContact{
			Name:  n.cfg.SenderName,
			Email: n.cfg.SenderEmail,
		},
		To: []brevoContact{{
			Name:  toName,
			Email: toEmail,
		}},
		Subject:     fmt.Sprintf("Oficina - Pos | OS %s atualizada", input.Code),
		HTMLContent: statusChangedHTML(input),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.cfg.APIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", n.cfg.APIKey)
	req.Header.Set("content-type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("brevo send email failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

type brevoEmailRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

type brevoContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func statusChangedHTML(input repository.StatusChangedNotification) string {
	code := html.EscapeString(input.Code)
	clientName := html.EscapeString(strings.TrimSpace(input.ClientName))
	if clientName == "" {
		clientName = "Cliente"
	}
	status := html.EscapeString(statusLabel(input.CurrentStatus))
	description := html.EscapeString(statusDescription(input.CurrentStatus))

	return fmt.Sprintf(`<!doctype html>
<html lang="pt-BR">
  <body style="margin:0;background:#f4f7fb;font-family:Arial,Helvetica,sans-serif;color:#182230;">
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f4f7fb;padding:32px 16px;">
      <tr>
        <td align="center">
          <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:640px;background:#ffffff;border-radius:16px;overflow:hidden;border:1px solid #e6edf5;">
            <tr>
              <td style="background:#101828;padding:28px 32px;color:#ffffff;">
                <div style="font-size:13px;letter-spacing:.08em;text-transform:uppercase;color:#9fd3ff;font-weight:700;">Oficina - Pos</div>
                <h1 style="margin:10px 0 0;font-size:26px;line-height:1.25;">Sua ordem de servico foi atualizada</h1>
              </td>
            </tr>
            <tr>
              <td style="padding:32px;">
                <p style="margin:0 0 18px;font-size:16px;line-height:1.6;">Ola, <strong>%s</strong>.</p>
                <p style="margin:0 0 24px;font-size:16px;line-height:1.6;">Temos uma atualizacao sobre a sua OS <strong>%s</strong>.</p>
                <div style="background:#eef7ff;border:1px solid #b9e0ff;border-radius:14px;padding:22px;margin:0 0 24px;">
                  <div style="font-size:13px;color:#175cd3;text-transform:uppercase;font-weight:700;margin-bottom:8px;">Status atual</div>
                  <div style="font-size:28px;line-height:1.2;font-weight:800;color:#101828;">%s</div>
                  <p style="margin:12px 0 0;font-size:15px;line-height:1.55;color:#344054;">%s</p>
                </div>
                <p style="margin:0;font-size:14px;line-height:1.6;color:#667085;">Voce recebera novas atualizacoes por e-mail conforme a oficina avancar nas proximas etapas.</p>
              </td>
            </tr>
            <tr>
              <td style="background:#f8fafc;padding:20px 32px;color:#667085;font-size:12px;line-height:1.5;">
                Este e um e-mail automatico da Oficina - Pos. Nao e necessario responder.
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, clientName, code, status, description)
}

func statusLabel(status string) string {
	switch status {
	case "RECEIVED":
		return "Recebida"
	case "IN_DIAGNOSIS":
		return "Diagnostico"
	case "WAITING_APPROVAL":
		return "Aguardando aprovacao"
	case "IN_PROGRESS":
		return "Em execucao"
	case "FINISHED":
		return "Finalizada"
	case "DELIVERED":
		return "Entregue"
	default:
		return status
	}
}

func statusDescription(status string) string {
	switch status {
	case "RECEIVED":
		return "Recebemos sua ordem de servico e ela ja esta registrada em nosso sistema."
	case "IN_DIAGNOSIS":
		return "Nossa equipe iniciou o diagnostico do veiculo para entender o que precisa ser feito."
	case "WAITING_APPROVAL":
		return "O orcamento esta pronto e aguardando a sua aprovacao."
	case "IN_PROGRESS":
		return "O servico foi aprovado e a execucao ja esta em andamento."
	case "FINISHED":
		return "O servico foi finalizado. Em breve voce recebera as orientacoes para retirada."
	case "DELIVERED":
		return "O veiculo foi entregue. Obrigado por escolher a Oficina Pos."
	default:
		return "A ordem de servico recebeu uma nova atualizacao."
	}
}
