package content

import "context"

type Form struct {
	FormID       string `json:"formId"`
	FormName     string `json:"formName"`
	FormFileID   string `json:"formFileId"`
	FormFileName string `json:"formFileName"`
}

type FormsResponse struct {
	Forms []Form `json:"forms"`
}

type News struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	NewsDate string `json:"newsDate"`
}

type NewsResponse struct {
	NewsGeneral []News `json:"newsGeneral"`
	NewsShip    []News `json:"newsShip"`
}

type Storage interface {
	Get(context.Context, string) ([]byte, string, error)
}
