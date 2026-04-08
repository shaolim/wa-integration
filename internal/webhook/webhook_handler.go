package webhook

import "net/http"

type Webhook struct {
}

func (w *Webhook) VerifyWebhook(rw http.ResponseWriter, r *http.Request) {

}
