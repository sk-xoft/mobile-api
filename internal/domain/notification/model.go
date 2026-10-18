package notification

type Item struct {
	ID             string `json:"id"`
	ReadStatus     string `json:"readStatus"`
	NotiType       string `json:"notiType"`
	MobileUserUUID string `json:"mobileUserUuid"`
	TitleMessage   string `json:"titleMessage"`
	BodyMessage    string `json:"bodyMessage"`
	Success        string `json:"success"`
	ReadDate       string `json:"readDate"`
	Retry          string `json:"retry"`
	NotiDate       string `json:"notiDate"`
	ValueID        string `json:"valueId"`
}
type Response struct {
	Notifications []Item `json:"notificationsModels"`
}
type CountResponse struct {
	Count string `json:"countNotification"`
}
type MarkRequest struct {
	NotiID string `json:"notiId"`
}
type MarkValueRequest struct {
	ValueID  string `json:"valueId"`
	NotiType string `json:"notiType"`
}
