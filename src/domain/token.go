package domain

type AccessDetails struct {
	TokenUuid string
	UserId    uint64
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type TemporaryTokenDetails struct {
	AccessToken string
	TokenUuid   string
	Expires     int64
}
