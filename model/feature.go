package model

type Feature struct {
	Id        string `json:"id"`        //package或bundle id
	Name      string `json:"name"`      //名称
	Icon      string `json:"icon"`      //图标；base64
	PublicKey string `json:"publicKey"` //公钥
	MD5       string `json:"md5"`       //签名
	Platform  string `json:"platform"`  //平台
}
