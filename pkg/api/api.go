package api

import (
	gohttp "net/http"

	"github.com/gorilla/pat"

	"github.com/doctolib/MailHog/pkg/config"
)

func CreateAPI(conf *config.Config, r gohttp.Handler) {
	apiv2 := createAPIv2(conf, r.(*pat.Router))

	go func() {
		for {
			select {
			case msg := <-conf.MessageChan:
				apiv2.messageChan <- msg
			}
		}
	}()
}
