package model

type Feature struct {
	Id        string //package或bundle id
	Name      string //名称
	Icon      string //图标；base64
	PublicKey string //公钥
	MD5       string //签名
}
