package dto

type ScanTotpDto struct {
	QRCode string `json:"qrcode"`
	Secret string `json:"secret"`
}
