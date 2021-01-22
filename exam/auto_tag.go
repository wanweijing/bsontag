package exam

type tagAAA struct {
	AAA string
	BBB string
	FFF struct {
		EEE   string
		FFF   string
		Inner struct {
			In     string
			Outter string
		}
	}
}
type tagTTTT struct {
	KKK struct {
		MMM   string
		Inner struct {
			In     string
			Outter string
		}
	}
	AAA string
}
type tagEx2 struct {
	Exxxx string
	Yyy   string
}
type tagEEEEEEFFFF2 struct {
	Exxxx string
	Yyy   string
}

var FN = struct {
	AAA         tagAAA
	TTTT        tagTTTT
	Ex2         tagEx2
	EEEEEEFFFF2 tagEEEEEEFFFF2
}{}

func init() {
	FN.AAA.AAA = "fuckddddddddd"
	FN.AAA.BBB = "fuck2"
	FN.AAA.FFF.EEE = "fff.eee"
	FN.AAA.FFF.FFF = "fff.fff"
	FN.AAA.FFF.Inner.In = "fff.inner.inner"
	FN.AAA.FFF.Inner.Outter = "fff.inner.outter"
	FN.TTTT.KKK.MMM = "kkk.mmm"
	FN.TTTT.KKK.Inner.In = "kkk.inner.inner"
	FN.TTTT.KKK.Inner.Outter = "kkk.inner.outter"
	FN.TTTT.AAA = "aaaa"
	FN.Ex2.Exxxx = "exxxx"
	FN.Ex2.Yyy = "yyy"
	FN.EEEEEEFFFF2.Exxxx = "exxxx"
	FN.EEEEEEFFFF2.Yyy = "yyy"
}
