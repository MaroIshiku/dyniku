package namecom

import (
	"net/http"

	"github.com/MaroIshiku/dyniku/internal/provider/headers"
)

func setHeaders(request *http.Request) {
	headers.SetAccept(request, "application/json")
	headers.SetUserAgent(request)
}
