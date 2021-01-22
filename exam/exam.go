package exam

type Inner struct {
	In     int `bson:"inner"`
	Outter int `bson:"outter"`
}

type DDD struct {
	EEE   int   `bson:"eee"`
	FFF   int   `bson:"fff"`
	Inner Inner `bson:"inner"`
}

// @table:yhyy
type AAA struct {
	AAA int    `json:"aaa" bson:"fuckddddddddd"`
	BBB string `json:"aaa" bson:"fuck2"`
	//CCC string // test 2

	DDD `bson:"fff"`
}

// @table:xxx
type TTTT struct {
	AAA int `bson:"aaaa"`
	KKK struct {
		MMM   int   `bson:"mmm"`
		Inner Inner `bson:"inner"`
	} `bson:"kkk"`
}
